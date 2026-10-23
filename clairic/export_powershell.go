package clairic

import (
	"fmt"
	"strconv"
	"strings"
)

// Single-quoted PowerShell strings preserve document text without interpolation.
func powerShellLiteral(value any) string {
	switch v := value.(type) {
	case string:
		quoted := strings.NewReplacer("'", "''", "‘", "‘‘", "’", "’’").Replace(v)
		return "'" + quoted + "'"
	case bool:
		if v {
			return "$true"
		}
		return "$false"
	case float64:
		return strconv.FormatFloat(v, 'g', -1, 64)
	default:
		panic("PowerShell export received an unvalidated field setting")
	}
}

func powerShellModuleProgram(request Request, schema Schema, embed bool) CodeExport {
	var code strings.Builder
	mode := "Reads a document file when run."
	if embed {
		mode = "Embeds the current document; an optional file argument overrides it."
	}
	code.WriteString("# " + mode + "\n")
	code.WriteString("# Credentials: CLOUDFLARE_ACCOUNT_ID and CLOUDFLARE_AUTH_TOKEN.\n")
	code.WriteString("# Install once if needed: Install-Module PSAIClairic -Scope CurrentUser\n")
	code.WriteString("# https://github.com/dfinke/PSAIClairic\n")
	code.WriteString("#Requires -Version 7.2\n")
	code.WriteString(conditional(embed, "param([string]$Path)\n\n", "param([Parameter(Mandatory)][string]$Path)\n\n"))
	code.WriteString("$ErrorActionPreference = 'Stop'\n")
	code.WriteString("Import-Module PSAIClairic -MinimumVersion 0.2.0 -ErrorAction Stop\n\n")
	code.WriteString("$apiToken = $env:CLOUDFLARE_AUTH_TOKEN\n")
	code.WriteString("if ([string]::IsNullOrWhiteSpace($apiToken)) {\n    $apiToken = $env:CLOUDFLARE_API_TOKEN\n}\n\n")
	if embed {
		fmt.Fprintf(&code, "$state = %s\n", powerShellLiteral(request.State))
		code.WriteString("if ($Path) {\n    $state = Get-Content -LiteralPath $Path -Raw -Encoding utf8\n}\n")
	} else {
		code.WriteString("$state = Get-Content -LiteralPath $Path -Raw -Encoding utf8\n")
	}
	code.WriteString("\n$fields = [ordered]@{\n")
	parameters := []struct{ key, name string }{
		{"instructions", "Instructions"}, {"unit", "Unit"}, {"integer", "Integer"},
		{"decimals", "Decimals"}, {"items", "Items"}, {"min", "Min"}, {"max", "Max"},
		{"step", "Step"}, {"scale", "Scale"}, {"null_threshold", "NullThreshold"},
		{"horizon", "Horizon"}, {"interval", "Interval"},
	}
	for _, field := range schema.Fields {
		spec := request.Fields[field.Name]
		fmt.Fprintf(&code, "    %s = New-ClairicField -Type %s", powerShellLiteral(field.Name), powerShellLiteral(spec["type"]))
		for _, parameter := range parameters {
			if value, ok := spec[parameter.key]; ok {
				separator := " "
				if parameter.key == "integer" {
					separator = ":"
				}
				fmt.Fprintf(&code, " -%s%s%s", parameter.name, separator, powerShellLiteral(value))
			}
		}
		code.WriteString("\n")
	}
	code.WriteString("}\n\n")
	code.WriteString("$result = Invoke-Clairic -State $state -Fields $fields -ReturnStateText -TimeoutSec 120 -ApiToken $apiToken")
	if request.ReferenceDate != "" {
		fmt.Fprintf(&code, " -ReferenceDate %s", powerShellLiteral(request.ReferenceDate))
	}
	if request.Options["include_quantiles"] {
		code.WriteString(" -IncludeQuantiles")
	}
	code.WriteString("\n$result\n\n")
	code.WriteString("# To use just the field values in a PowerShell pipeline:\n")
	code.WriteString("# $result | ConvertFrom-ClairicAnswer -FieldOrder $fields.Keys\n")
	return CodeExport{
		Language: "PowerShell", Filename: "Extract-Document.ps1", Code: code.String(),
		Note: "PowerShell 7.2+ with PSAIClairic 0.2.0+. Uses New-ClairicField and Invoke-Clairic; retains full answers and metadata. Credentials come from your environment.",
	}
}
