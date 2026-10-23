package clairic

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTTPUnicodeEnvelope(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer secret" {
			t.Error("missing auth")
		}
		body, _ := io.ReadAll(r.Body)
		var request Request
		if json.Unmarshal(body, &request) != nil || request.State != "Café 😀" {
			t.Error("UTF-8 request changed")
		}
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"success":true,"result":{"answers":{"invoice":{"value":"Café","confidence":0.99,"span":[0,4]}},"truncated":false}}`)
	}))
	defer server.Close()
	c := NewClient()
	c.BaseURL = server.URL
	r, _ := BuildRequest("Café 😀", sample())
	result, e := c.Extract(context.Background(), strings.Repeat("a", 32), "secret", r)
	if e != nil || len(result.Answers) != 1 {
		t.Fatalf("%+v %v", result, e)
	}
}
func TestHTTPFailuresDoNotEchoSecrets(t *testing.T) {
	for _, status := range []int{401, 429, 500} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(status)
			io.WriteString(w, "SECRET-TOKEN document contents")
		}))
		c := NewClient()
		c.BaseURL = server.URL
		r, _ := BuildRequest("x", sample())
		_, e := c.Extract(context.Background(), strings.Repeat("a", 32), "SECRET-TOKEN", r)
		server.Close()
		if e == nil || strings.Contains(e.Error(), "SECRET") || strings.Contains(e.Error(), "document contents") {
			t.Fatalf("unsafe error: %v", e)
		}
	}
}
func TestMalformedMissingAndDirectResponse(t *testing.T) {
	for _, body := range []string{`not JSON`, `{"success":false}`, `{"success":true}`, `{"answers":{"other":{"value":1}}}`, `{"answers":{"invoice":{"guess":"wrong"}}}`} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { io.WriteString(w, body) }))
		c := NewClient()
		c.BaseURL = server.URL
		req, _ := BuildRequest("x", sample())
		_, e := c.Extract(context.Background(), strings.Repeat("a", 32), "token", req)
		server.Close()
		if e == nil {
			t.Fatalf("malformed response accepted: %s", body)
		}
	}
}
func TestNullIsNotReplacedByGuess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, `{"answers":{"invoice":{"value":null,"guess":"not the value"}}}`)
	}))
	defer server.Close()
	c := NewClient()
	c.BaseURL = server.URL
	req, _ := BuildRequest("x", sample())
	result, e := c.Extract(context.Background(), strings.Repeat("a", 32), "token", req)
	if e != nil {
		t.Fatal(e)
	}
	var value map[string]any
	json.Unmarshal(result.Answers["invoice"], &value)
	if value["value"] != nil {
		t.Fatal("missingness changed")
	}
}
