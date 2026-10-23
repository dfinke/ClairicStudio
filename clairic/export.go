package clairic

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"go/format"
	"strconv"
	"strings"
)

type CodeExport struct {
	Language string `json:"language"`
	Filename string `json:"filename"`
	Code     string `json:"code"`
	Note     string `json:"note"`
}

func ExportProgram(language, state string, s Schema, embed bool) (CodeExport, error) {
	return ExportProgramWithStyle(language, state, s, embed, "standalone")
}

func ExportProgramWithStyle(language, state string, s Schema, embed bool, style string) (CodeExport, error) {
	if style == "" {
		style = "standalone"
	}
	if style != "standalone" && style != "module" {
		return CodeExport{}, errors.New("Choose standalone or module export style")
	}
	if style == "module" && language != "PowerShell" {
		return CodeExport{}, errors.New("The PSAIClairic module export is available for PowerShell only")
	}
	validationState := state
	if !embed && strings.TrimSpace(state) == "" {
		validationState = "Document supplied at runtime."
	}
	r, err := BuildRequest(validationState, s)
	if err != nil {
		return CodeExport{}, err
	}
	if !embed {
		r.State = ""
	}
	if style == "module" {
		return powerShellModuleProgram(r, s, embed), nil
	}
	payload, _ := json.Marshal(r)
	encoded := base64.StdEncoding.EncodeToString(payload)
	e := CodeExport{Language: language}
	// A base64-encoded JSON literal preserves quotes, Unicode, and line breaks in every target.
	// It is only request data, never evaluated as source code, and never contains credentials.
	switch language {
	case "Python":
		e.Filename = "extract.py"
		e.Note = "Python 3; standard library only."
		e.Code = `import base64, json, os, sys, urllib.request
request = json.loads(base64.b64decode("` + encoded + `"))
` + conditional(embed, `if len(sys.argv) > 1:
    request["state"] = open(sys.argv[1], encoding="utf-8-sig").read()
`, `if len(sys.argv) != 2:
    raise SystemExit("Usage: python extract.py document.txt")
request["state"] = open(sys.argv[1], encoding="utf-8-sig").read()
`) + `account = os.environ["CLOUDFLARE_ACCOUNT_ID"]
token = os.environ.get("CLOUDFLARE_AUTH_TOKEN", "").strip() or os.environ.get("CLOUDFLARE_API_TOKEN", "").strip()
url = f"https://api.cloudflare.com/client/v4/accounts/{account}/ai/run/@cf/cloudflare/clairic"
req = urllib.request.Request(url, json.dumps(request, ensure_ascii=False).encode("utf-8"),
    {"Authorization": f"Bearer {token}", "Content-Type": "application/json"})
with urllib.request.urlopen(req, timeout=120) as response:
    result = json.load(response)
if result.get("success") is False:
    raise RuntimeError("Cloudflare reported an unsuccessful request")
print(json.dumps(result.get("result", result), ensure_ascii=False, indent=2))
`
	case "PowerShell":
		e.Filename = "Extract-Document.ps1"
		e.Note = "PowerShell 7.2+; calls the HTTP API directly, with no module dependency."
		e.Code = "#Requires -Version 7.2\n" + conditional(embed, "param([string]$Path)\n", "param([Parameter(Mandatory)][string]$Path)\n") + `$ErrorActionPreference = 'Stop'
$request = [Text.Encoding]::UTF8.GetString([Convert]::FromBase64String('` + encoded + `')) | ConvertFrom-Json -AsHashtable
if ($Path) {
    $request.state = Get-Content -LiteralPath $Path -Raw -Encoding utf8
}
$account = $env:CLOUDFLARE_ACCOUNT_ID
$token = $env:CLOUDFLARE_AUTH_TOKEN
if ([string]::IsNullOrWhiteSpace($token)) {
    $token = $env:CLOUDFLARE_API_TOKEN
}
if (-not $account -or -not $token) {
    throw 'Set CLOUDFLARE_ACCOUNT_ID and CLOUDFLARE_AUTH_TOKEN.'
}
$body = [Text.Encoding]::UTF8.GetBytes(($request | ConvertTo-Json -Depth 60 -Compress))
$response = Invoke-RestMethod -Method Post -Uri "https://api.cloudflare.com/client/v4/accounts/$account/ai/run/@cf/cloudflare/clairic" -Headers @{Authorization="Bearer $token"} -ContentType 'application/json; charset=utf-8' -Body $body -TimeoutSec 120
if ($response.PSObject.Properties['success'] -and -not $response.success) {
    throw 'Cloudflare reported an unsuccessful request.'
}
if ($response.PSObject.Properties['result']) {
    $response.result
} else {
    $response
}
`
	case "TypeScript":
		e.Filename = "extract.ts"
		e.Note = "Node.js 18+ with a TypeScript runner, e.g. npx tsx extract.ts [document.txt]."
		e.Code = `import { readFileSync } from 'node:fs';
const request = JSON.parse(Buffer.from('` + encoded + `', 'base64').toString('utf8'));
` + conditional(embed, `if (process.argv[2]) {
  request.state = readFileSync(process.argv[2], 'utf8').replace(/^\uFEFF/, '');
}
`, `if (!process.argv[2]) {
  throw new Error('Usage: npx tsx extract.ts document.txt');
}
request.state = readFileSync(process.argv[2], 'utf8').replace(/^\uFEFF/, '');
`) + `const account = process.env.CLOUDFLARE_ACCOUNT_ID;
const token = process.env.CLOUDFLARE_AUTH_TOKEN?.trim() || process.env.CLOUDFLARE_API_TOKEN?.trim();
if (!account || !token) {
  throw new Error('Set CLOUDFLARE_ACCOUNT_ID and CLOUDFLARE_AUTH_TOKEN');
}
const response = await fetch(` + "`https://api.cloudflare.com/client/v4/accounts/${account}/ai/run/@cf/cloudflare/clairic`" + `, {
  method: 'POST', headers: { Authorization: ` + "`Bearer ${token}`" + `, 'Content-Type': 'application/json' },
  body: JSON.stringify(request), signal: AbortSignal.timeout(120000),
});
if (!response.ok) {
  throw new Error('Cloudflare HTTP ' + response.status);
}
const result = await response.json();
if (result.success === false) {
  throw new Error('Cloudflare reported an unsuccessful request');
}
console.log(JSON.stringify(result.result ?? result, null, 2));
`
	case "Go":
		e.Filename = "extract.go"
		e.Note = "Go 1.22+; standard library only. Run go run extract.go [document.txt]."
		e.Code = `package main
import ("bytes"; "encoding/base64"; "encoding/json"; "fmt"; "io"; "net/http"; "os"; "strings"; "time")
func main() {
 data, _ := base64.StdEncoding.DecodeString("` + encoded + `")
 var request map[string]any
 if err := json.Unmarshal(data, &request); err != nil {
  panic(err)
 }
` + conditional(embed, "", ` if len(os.Args) != 2 {
  panic("Usage: go run extract.go document.txt")
 }
`) + ` if len(os.Args) > 1 {
  text, err := os.ReadFile(os.Args[1])
  if err != nil {
   panic(err)
  }
  request["state"] = strings.TrimPrefix(string(text), "\ufeff")
 }
 account, token := os.Getenv("CLOUDFLARE_ACCOUNT_ID"), os.Getenv("CLOUDFLARE_AUTH_TOKEN")
 if strings.TrimSpace(token) == "" {
  token = os.Getenv("CLOUDFLARE_API_TOKEN")
 }
 if account == "" || token == "" {
  panic("Set CLOUDFLARE_ACCOUNT_ID and CLOUDFLARE_AUTH_TOKEN")
 }
 body, err := json.Marshal(request)
 if err != nil {
  panic(err)
 }
 req, err := http.NewRequest("POST", "https://api.cloudflare.com/client/v4/accounts/"+account+"/ai/run/@cf/cloudflare/clairic", bytes.NewReader(body))
 if err != nil {
  panic(err)
 }
 req.Header.Set("Authorization", "Bearer "+token)
 req.Header.Set("Content-Type", "application/json")
 response, err := (&http.Client{Timeout:120*time.Second}).Do(req)
 if err != nil {
  panic("Cloudflare request failed")
 }
 defer response.Body.Close()
 if response.StatusCode < 200 || response.StatusCode >= 300 {
  panic(fmt.Sprintf("Cloudflare HTTP %d", response.StatusCode))
 }
 data, err = io.ReadAll(response.Body)
 if err != nil {
  panic(err)
 }
 var result map[string]json.RawMessage
 if err=json.Unmarshal(data, &result); err != nil {
  panic(err)
 }
 if string(result["success"]) == "false" {
  panic("Cloudflare reported an unsuccessful request")
 }
 if raw, ok := result["result"]; ok {
  data=raw
 }
 var pretty bytes.Buffer
 if err=json.Indent(&pretty,data,"","  "); err != nil {
  panic(err)
 }
 fmt.Println(pretty.String())
}
`
	case "Rust":
		e.Filename = "main.rs"
		e.Note = "Cargo project required: add reqwest 0.12 (blocking,json), serde_json 1, and base64 0.22 to Cargo.toml."
		e.Code = `// Save as src/main.rs in a Cargo project (edition = "2021").
// [dependencies]
// reqwest = { version = "0.12", features = ["blocking", "json"] }
// serde_json = "1"
// base64 = "0.22"
use base64::Engine;
fn main() -> Result<(), Box<dyn std::error::Error>> {
 let data = base64::engine::general_purpose::STANDARD.decode("` + encoded + `")?;
 let mut request: serde_json::Value = serde_json::from_slice(&data)?;
 let path = std::env::args().nth(1);
` + conditional(embed, "", ` if path.is_none() {
     return Err("Usage: cargo run -- document.txt".into());
 }
`) + ` if let Some(path) = path {
     request["state"] = std::fs::read_to_string(path)?.trim_start_matches('\u{feff}').into();
 }
 let account = std::env::var("CLOUDFLARE_ACCOUNT_ID")?;
 let token = std::env::var("CLOUDFLARE_AUTH_TOKEN").ok()
  .filter(|value| !value.trim().is_empty())
  .or_else(|| std::env::var("CLOUDFLARE_API_TOKEN").ok())
  .ok_or("Set CLOUDFLARE_AUTH_TOKEN or CLOUDFLARE_API_TOKEN")?;
 let response = reqwest::blocking::Client::builder().timeout(std::time::Duration::from_secs(120)).build()?
  .post(format!("https://api.cloudflare.com/client/v4/accounts/{account}/ai/run/@cf/cloudflare/clairic"))
  .bearer_auth(token).json(&request).send()?.error_for_status()?;
 let value: serde_json::Value = response.json()?;
 if value.get("success").and_then(|v|v.as_bool()) == Some(false) {
     return Err("Cloudflare reported an unsuccessful request".into());
 }
 println!("{}", serde_json::to_string_pretty(value.get("result").unwrap_or(&value))?);
 Ok(())
}
`
	default:
		return e, errors.New("Choose Python, PowerShell, TypeScript, Go, or Rust")
	}
	pretty, _ := json.MarshalIndent(r, "", "  ")
	quoted, _ := json.Marshal(string(pretty))
	switch language {
	case "Python":
		e.Code = strings.Replace(e.Code, "import base64, json, os, sys, urllib.request", "import json, os, sys, urllib.request", 1)
		e.Code = strings.Replace(e.Code, `request = json.loads(base64.b64decode("`+encoded+`"))`, `request = json.loads(`+string(quoted)+`)`, 1)
	case "PowerShell":
		e.Code = strings.Replace(e.Code, `$request = [Text.Encoding]::UTF8.GetString([Convert]::FromBase64String('`+encoded+`')) | ConvertFrom-Json -AsHashtable`, "$request = "+powerShellLiteral(string(pretty))+" | ConvertFrom-Json -AsHashtable", 1)
	case "TypeScript":
		e.Code = strings.Replace(e.Code, `const request = JSON.parse(Buffer.from('`+encoded+`', 'base64').toString('utf8'));`, `const request = JSON.parse(`+string(quoted)+`);`, 1)
		e.Code = strings.ReplaceAll(e.Code, `/^\\uFEFF/`, `/^\uFEFF/`)
	case "Go":
		e.Code = strings.Replace(e.Code, `"encoding/base64"; `, "", 1)
		e.Code = strings.Replace(e.Code, `data, _ := base64.StdEncoding.DecodeString("`+encoded+`")`, `data := []byte(`+strconv.Quote(string(pretty))+`)`, 1)
		formatted, err := format.Source([]byte(e.Code))
		if err != nil {
			return e, fmt.Errorf("Format Go export: %w", err)
		}
		e.Code = string(formatted)
	case "Rust":
		hashes := "#"
		for strings.Contains(string(pretty), `"`+hashes) {
			hashes += "#"
		}
		e.Code = strings.Replace(e.Code, "use base64::Engine;\n", "", 1)
		start := ` let data = base64::engine::general_purpose::STANDARD.decode("` + encoded + `")?;` + "\n" + ` let mut request: serde_json::Value = serde_json::from_slice(&data)?;`
		e.Code = strings.Replace(e.Code, start, ` let mut request: serde_json::Value = serde_json::from_str(r`+hashes+`"`+string(pretty)+`"`+hashes+`)?;`, 1)
		e.Code = strings.Replace(e.Code, "// base64 = \"0.22\"\n", "", 1)
		e.Note = "Cargo project required: add reqwest 0.12 (blocking,json) and serde_json 1 to Cargo.toml."
	}
	mode := "Reads a document file when run."
	if embed {
		mode = "Embeds the current document; an optional file argument overrides it."
	}
	e.Code = comment(language, mode+" Credentials are read from environment variables.") + e.Code
	return e, nil
}
func conditional(b bool, a, c string) string {
	if b {
		return a
	}
	return c
}
func comment(lang, text string) string {
	if contains(lang, "Python", "PowerShell") {
		return "# " + text + "\n"
	}
	return fmt.Sprintf("// %s\n", strings.TrimSpace(text))
}
