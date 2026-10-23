package clairic

import (
	"encoding/json"
	"strings"
	"testing"
)

func sample() Schema {
	return Schema{Version: 1, Fields: []Field{{Name: "invoice", Type: "text", Settings: map[string]any{}}}}
}
func TestAllTypes(t *testing.T) {
	for _, kind := range []string{"text", "number", "date", "time", "datetime", "duration", "list", "estimate", "forecast"} {
		t.Run(kind, func(t *testing.T) {
			f := Field{Name: "field", Type: kind, Settings: map[string]any{}}
			switch kind {
			case "list":
				f.Settings["items"] = "text"
			case "estimate", "forecast":
				f.Settings["min"] = "0"
				f.Settings["max"] = "100"
				f.Settings["step"] = "1"
			}
			s := Schema{Version: 1, Fields: []Field{f}}
			r, e := BuildRequest("Café's invoice — €12 😀", s)
			if e != nil {
				t.Fatal(e)
			}
			if r.Fields["field"]["type"] != kind {
				t.Fatal("lost type")
			}
		})
	}
}
func TestSemanticValidation(t *testing.T) {
	cases := []struct {
		name  string
		field Field
	}{
		{"name", Field{Name: "bad name", Type: "text"}},
		{"list", Field{Name: "x", Type: "list"}},
		{"precision", Field{Name: "x", Type: "number", Settings: map[string]any{"decimals": "7"}}},
		{"null", Field{Name: "x", Type: "text", Settings: map[string]any{"null_threshold": "0"}}},
		{"bounds", Field{Name: "x", Type: "estimate", Settings: map[string]any{"min": 10, "max": 1}}},
		{"log", Field{Name: "x", Type: "estimate", Settings: map[string]any{"min": 0, "max": 10, "scale": "log"}}},
		{"finite", Field{Name: "x", Type: "number", Settings: map[string]any{"decimals": "NaN"}}},
		{"nested", Field{Name: "x", Type: "number", Settings: map[string]any{"unit": map[string]any{"oops": true}}}},
		{"duration", Field{Name: "x", Type: "duration", Settings: map[string]any{"unit": "furlongs"}}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s := sample()
			s.Fields = []Field{c.field}
			if _, e := BuildRequest("document", s); e == nil {
				t.Fatal("invalid field accepted")
			}
		})
	}
}
func TestDuplicateAndDateBounds(t *testing.T) {
	s := sample()
	s.Fields = append(s.Fields, Field{Name: "INVOICE", Type: "text"})
	if _, e := BuildRequest("x", s); e == nil {
		t.Fatal("duplicate accepted")
	}
	s.Fields = []Field{{Name: "arrival", Type: "estimate", Settings: map[string]any{"min": "2026-10-01", "max": "2026-12-01", "step": "month"}}}
	if _, e := BuildRequest("x", s); e != nil {
		t.Fatal(e)
	}
	s.ReferenceDate = "2026-02-30"
	if _, e := BuildRequest("x", s); e == nil {
		t.Fatal("bad date accepted")
	}
}
func TestPowerShellSchemaAndIncompleteDraft(t *testing.T) {
	data := []byte(`{"version":1,"fields":[{"name":"unfinished name","type":"number","instructions":"Café","settings":{"decimals":"2"}}]}`)
	s, e := ReadSchema(data)
	if e != nil {
		t.Fatal(e)
	}
	s.Fields[0].Name = "total"
	r, e := BuildRequest("x", s)
	if e != nil {
		t.Fatal(e)
	}
	if r.Fields["total"]["decimals"] != float64(2) {
		t.Fatal("string setting not translated")
	}
	encoded, _ := json.Marshal(s)
	if _, e = ReadSchema(encoded); e != nil {
		t.Fatal(e)
	}
	if _, e = ReadSchema([]byte(`{"version":2,"fields":[]}`)); e == nil {
		t.Fatal("unknown version accepted")
	}
}
func TestTextListOmitsNumericPrecision(t *testing.T) {
	s := sample()
	s.Fields = []Field{{Name: "names", Type: "list", Settings: map[string]any{"items": "text", "decimals": "2"}}}
	r, e := BuildRequest("x", s)
	if e != nil {
		t.Fatal(e)
	}
	if _, ok := r.Fields["names"]["decimals"]; ok {
		t.Fatal("irrelevant precision leaked")
	}
}
func TestDocumentValidation(t *testing.T) {
	for _, text := range []string{"", " ", "binary\x00text", strings.Repeat("x", 8<<20+1)} {
		if _, e := BuildRequest(text, sample()); e == nil {
			t.Fatal("invalid document accepted")
		}
	}
}
