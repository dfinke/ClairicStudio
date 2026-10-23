package clairic

import (
	"strings"
	"testing"
)

func TestPowerShellModuleExportModes(t *testing.T) {
	schema := sample()
	schema.ReferenceDate = "2026-10-23"
	schema.IncludeQuantiles = true
	for _, embedded := range []bool{false, true} {
		exported, err := ExportProgramWithStyle("PowerShell", "Invoice 'quoted' 😀", schema, embedded, "module")
		if err != nil {
			t.Fatal(err)
		}
		for _, required := range []string{"Import-Module PSAIClairic", "New-ClairicField -Type 'text'", "[ordered]@{", "Invoke-Clairic -State $state -Fields $fields", "-ReferenceDate '2026-10-23'", "-IncludeQuantiles", "-ReturnStateText", "ConvertFrom-ClairicAnswer -FieldOrder $fields.Keys"} {
			if !strings.Contains(exported.Code, required) {
				t.Fatalf("missing %s", required)
			}
		}
		if strings.Contains(exported.Code, "Invoke-RestMethod") || !strings.Contains(exported.Note, "PSAIClairic") {
			t.Fatal("module export did not use the module")
		}
		if strings.Contains(exported.Code, "Invoice ''quoted'' 😀") != embedded {
			t.Fatal("document input mode was not respected")
		}
		if strings.Contains(exported.Code, "[Parameter(Mandatory)]") == embedded {
			t.Fatal("incorrect path requirement")
		}
	}
}

func TestModuleExportOptionsAreValidated(t *testing.T) {
	for _, c := range []struct{ language, style string }{{"Python", "module"}, {"PowerShell", "unknown"}} {
		if _, err := ExportProgramWithStyle(c.language, "invoice", sample(), true, c.style); err == nil {
			t.Fatal("unsupported export style accepted")
		}
	}
	legacy, err := ExportProgram("PowerShell", "invoice", sample(), true)
	if err != nil {
		t.Fatal(err)
	}
	unchanged, err := ExportProgramWithStyle("PowerShell", "invoice", sample(), true, "")
	if err != nil || legacy.Code != unchanged.Code || strings.Contains(unchanged.Code, "Import-Module PSAIClairic") {
		t.Fatal("default export changed")
	}
}

func TestModuleSettingsUseValidatedTypesAndKeepOrder(t *testing.T) {
	schema := Schema{Version: 1, Fields: []Field{
		{Name: "units", Type: "number", Settings: map[string]any{"integer": false, "decimals": "2", "null_threshold": "0.5"}},
		{Name: "invoice", Type: "text", Instructions: "Customer's ‘quoted’ value", Settings: map[string]any{}},
	}}
	exported, err := ExportProgramWithStyle("PowerShell", "invoice", schema, true, "module")
	if err != nil {
		t.Fatal(err)
	}
	for _, required := range []string{"-Integer:$false", "-Decimals 2", "-NullThreshold 0.5", "Customer''s ‘‘quoted’’ value"} {
		if !strings.Contains(exported.Code, required) {
			t.Fatalf("missing typed setting %s", required)
		}
	}
	if strings.Index(exported.Code, "'units' =") > strings.Index(exported.Code, "'invoice' =") {
		t.Fatal("schema field order changed")
	}
}
