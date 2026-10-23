// Package clairic provides UI-independent Clairic request validation and transport.
package clairic

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

type Field struct {
	Name         string         `json:"name"`
	Type         string         `json:"type"`
	Instructions string         `json:"instructions"`
	Settings     map[string]any `json:"settings"`
}
type Schema struct {
	Version          int     `json:"version"`
	Fields           []Field `json:"fields"`
	ReferenceDate    string  `json:"reference_date,omitempty"`
	IncludeQuantiles bool    `json:"include_quantiles,omitempty"`
}
type Request struct {
	Model         string                    `json:"model"`
	State         string                    `json:"state"`
	Fields        map[string]map[string]any `json:"fields"`
	ReferenceDate string                    `json:"reference_date,omitempty"`
	Options       map[string]bool           `json:"options"`
}

var fieldName = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_.\-]{0,63}$`)
var AccountPattern = regexp.MustCompile(`^[a-fA-F0-9]{32}$`)
var types = map[string]bool{"text": true, "number": true, "date": true, "time": true, "datetime": true, "duration": true, "list": true, "estimate": true, "forecast": true}
var settings = map[string]bool{"unit": true, "integer": true, "decimals": true, "items": true, "min": true, "max": true, "step": true, "scale": true, "null_threshold": true, "horizon": true, "interval": true}

// ReadSchema accepts the existing version 1 PowerShell designer format, including unfinished fields.
func ReadSchema(data []byte) (Schema, error) {
	var schema Schema
	if len(data) > 2<<20 {
		return schema, errors.New("Schema files must be no larger than 2 MiB")
	}
	if !utf8.Valid(data) {
		return schema, errors.New("Schema must be UTF-8")
	}
	data = []byte(strings.TrimPrefix(string(data), "\ufeff"))
	if err := json.Unmarshal(data, &schema); err != nil {
		return schema, errors.New("Choose a valid schema JSON file")
	}
	if schema.Version != 1 || schema.Fields == nil || len(schema.Fields) > 500 {
		return schema, errors.New("Expected a version 1 schema with a fields array (up to 500 fields)")
	}
	for _, f := range schema.Fields {
		if !types[f.Type] {
			return schema, fmt.Errorf("Unknown field type %q", f.Type)
		}
		for k, v := range f.Settings {
			if !settings[k] {
				return schema, fmt.Errorf("Unknown setting %q", k)
			}
			switch v.(type) {
			case map[string]any, []any:
				return schema, fmt.Errorf("Setting %s must be a scalar value", k)
			}
		}
	}
	return schema, nil
}

func number(v any) (float64, error) {
	var n float64
	var err error
	switch x := v.(type) {
	case float64:
		n = x
	case int:
		n = float64(x)
	case json.Number:
		n, err = x.Float64()
	case string:
		n, err = strconv.ParseFloat(x, 64)
	default:
		return 0, errors.New("use a number")
	}
	if err != nil || math.IsNaN(n) || math.IsInf(n, 0) {
		return 0, errors.New("use a finite number")
	}
	return n, nil
}
func contains(v string, choices ...string) bool {
	for _, c := range choices {
		if v == c {
			return true
		}
	}
	return false
}
func allowed(t, items string, k string) bool {
	switch k {
	case "null_threshold":
		return contains(t, "text", "number", "date", "time", "datetime", "duration")
	case "unit":
		return contains(t, "number", "duration", "estimate", "forecast") || t == "list" && items == "number"
	case "integer":
		return contains(t, "number", "duration")
	case "decimals":
		return contains(t, "number", "duration") || t == "list" && items == "number"
	case "items":
		return t == "list"
	case "min", "max", "step", "scale":
		return contains(t, "estimate", "forecast")
	case "horizon", "interval":
		return t == "forecast"
	}
	return false
}
func BuildRequest(state string, s Schema) (Request, error) {
	r := Request{Model: "clairic", State: state, Fields: map[string]map[string]any{}, ReferenceDate: s.ReferenceDate, Options: map[string]bool{"return_state_text": true, "include_quantiles": s.IncludeQuantiles}}
	if s.Version != 1 {
		return r, errors.New("Use a version 1 schema")
	}
	if strings.TrimSpace(state) == "" {
		return r, errors.New("Paste or open a document first")
	}
	if len(state) > 8<<20 || !utf8.ValidString(state) || strings.ContainsRune(state, 0) {
		return r, errors.New("Document must be UTF-8 plain text no larger than 8 MiB")
	}
	if len(s.Fields) < 1 || len(s.Fields) > 500 {
		return r, errors.New("Add between 1 and 500 fields")
	}
	if s.ReferenceDate != "" {
		if _, err := time.Parse("2006-01-02", s.ReferenceDate); err != nil {
			return r, errors.New("Reference date must be YYYY-MM-DD")
		}
	}
	seen := map[string]bool{}
	for _, f := range s.Fields {
		fail := func(msg string) (Request, error) { return r, fmt.Errorf("%s: %s", f.Name, msg) }
		if !fieldName.MatchString(f.Name) {
			return fail("use 1–64 letters, digits, underscores, dots or hyphens; start with a letter or underscore")
		}
		lower := strings.ToLower(f.Name)
		if seen[lower] {
			return fail("field names must be unique")
		}
		seen[lower] = true
		if !types[f.Type] {
			return fail("choose a supported field type")
		}
		spec := map[string]any{"type": f.Type}
		if f.Instructions != "" {
			spec["instructions"] = f.Instructions
		}
		items, _ := f.Settings["items"].(string)
		for k, v := range f.Settings {
			if !settings[k] {
				return fail("unknown setting " + k)
			}
			empty, isText := v.(string)
			if v == nil || (isText && empty == "") || !allowed(f.Type, items, k) {
				continue
			}
			switch k {
			case "integer":
				if _, ok := v.(bool); !ok {
					return fail("integer must be true or false")
				}
				spec[k] = v
			case "decimals", "horizon", "null_threshold":
				n, err := number(v)
				if err != nil {
					return fail(k + ": " + err.Error())
				}
				if k == "decimals" && (n < 0 || n > 6 || n != math.Trunc(n)) {
					return fail("decimals must be a whole number from 0 to 6")
				}
				if k == "horizon" && (n < 1 || n > 64 || n != math.Trunc(n)) {
					return fail("horizon must be a whole number from 1 to 64")
				}
				if k == "null_threshold" && (n <= 0 || n > 1) {
					return fail("null threshold must be greater than 0 and at most 1")
				}
				spec[k] = n
			case "min", "max":
				n, err := number(v)
				if err == nil {
					spec[k] = n
				} else if text, ok := v.(string); ok {
					if _, e := time.Parse("2006-01-02", text); e != nil {
						return fail(k + " must be a number or ISO date")
					}
					spec[k] = text
				} else {
					return fail(k + " must be a number or ISO date")
				}
			case "step":
				if text, ok := v.(string); ok && contains(text, "month", "year") {
					spec[k] = text
				} else {
					n, err := number(v)
					if err != nil || n <= 0 {
						return fail("step must be positive, month, or year")
					}
					spec[k] = n
				}
			default:
				text, ok := v.(string)
				if !ok {
					return fail(k + " must be text")
				}
				spec[k] = text
			}
		}
		if f.Type == "list" && !contains(items, "text", "number", "date") {
			return fail("choose list item type text, number, or date")
		}
		if unit, ok := spec["unit"].(string); ok && f.Type == "duration" && !contains(unit, "seconds", "minutes", "hours", "days", "weeks", "months", "years") {
			return fail("choose a supported duration unit")
		}
		if scale, ok := spec["scale"].(string); ok && !contains(scale, "linear", "log") {
			return fail("scale must be linear or log")
		}
		if interval, ok := spec["interval"].(string); ok && !contains(interval, "step", "minute", "hour", "day", "week", "month", "quarter", "year") {
			return fail("choose a supported forecast interval")
		}
		if contains(f.Type, "estimate", "forecast") {
			lo, loOK := spec["min"]
			hi, hiOK := spec["max"]
			if !loOK || !hiOK {
				return fail("min and max are required")
			}
			if a, ok := lo.(float64); ok {
				b, ok := hi.(float64)
				if !ok || a >= b {
					return fail("require min < max with matching bound types")
				}
				if spec["scale"] == "log" && a <= 0 {
					return fail("log scale requires min > 0")
				}
			} else {
				a, aOK := lo.(string)
				b, bOK := hi.(string)
				if !aOK || !bOK || a >= b {
					return fail("require min < max with matching bound types")
				}
				if f.Type == "forecast" {
					return fail("date bounds are supported for estimates, not forecasts")
				}
				if spec["scale"] == "log" {
					return fail("log scale requires numeric bounds")
				}
			}
		}
		r.Fields[f.Name] = spec
	}
	encoded, err := json.Marshal(r)
	if err != nil {
		return r, errors.New("Request contains an unsupported setting")
	}
	if len(encoded) > 16<<20 {
		return r, errors.New("Request exceeds 16 MiB")
	}
	return r, nil
}
