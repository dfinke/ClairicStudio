package clairic

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBundledSamplesBuildRequests(t *testing.T) {
	for _, name := range []string{"invoice", "all-types"} {
		data, err := os.ReadFile(filepath.Join("..", "samples", name+".schema.json"))
		if err != nil {
			t.Fatal(err)
		}
		schema, err := ReadSchema(data)
		if err != nil {
			t.Fatal(err)
		}
		document, err := os.ReadFile(filepath.Join("..", "samples", name+".txt"))
		if err != nil {
			t.Fatal(err)
		}
		if _, err = BuildRequest(string(document), schema); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
	}
}

func TestNestedSettingFailsWithoutPanic(t *testing.T) {
	s := Schema{Version: 1, Fields: []Field{{Name: "amount", Type: "number", Settings: map[string]any{"decimals": map[string]any{"bad": true}}}}}
	if _, err := BuildRequest("USD 10", s); err == nil {
		t.Fatal("nested numeric setting accepted")
	}
}
