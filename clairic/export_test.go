package clairic

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

func TestExportsKeepDefinitionAndSecretsOut(t *testing.T) {
	state := "Invoice 'quoted'\n'@\n${not_code} `ticks` 😀 Café \\\n"
	for _, lang := range []string{"PowerShell", "Python", "TypeScript", "Go", "Rust"} {
		for _, embedded := range []bool{false, true} {
			e, err := ExportProgram(lang, state, sample(), embedded)
			if err != nil {
				t.Fatal(err)
			}
			if e.Filename == "" || !strings.Contains(e.Code, "CLOUDFLARE_AUTH_TOKEN") || !strings.Contains(e.Code, "invoice") {
				t.Fatalf("incomplete %s export", lang)
			}
			if !embedded && strings.Contains(e.Code, "quoted") {
				t.Fatal("file export retained document")
			}
			if lang == "Go" {
				if _, err = parser.ParseFile(token.NewFileSet(), e.Filename, e.Code, parser.AllErrors); err != nil {
					t.Fatal(err)
				}
			}
		}
	}
}
func TestFileExportWithoutCurrentDocument(t *testing.T) {
	if _, e := ExportProgram("Python", "", sample(), false); e != nil {
		t.Fatal(e)
	}
	if _, e := ExportProgram("Python", "", sample(), true); e == nil {
		t.Fatal("empty embedded document accepted")
	}
}
func TestRustDelimiterAndPrototypeField(t *testing.T) {
	s := sample()
	s.Fields[0].Name = "__proto__"
	e, err := ExportProgram("Rust", `text "# and "## and "###`, s, true)
	if err != nil || !strings.Contains(e.Code, `r####"`) {
		t.Fatalf("unsafe raw delimiter: %v\n%s", err, e.Code)
	}
}
