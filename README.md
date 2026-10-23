<p align="center"><img src="build/appicon.png" width="96" alt="Clairic Studio icon"></p>

# Clairic Studio

**Turn messy text into typed data—and the code to do it again.**

Paste an invoice. Name the fields you want. Watch the results take shape as you refine the schema. Then copy a working program in **PowerShell, Python, TypeScript, Go, or Rust** and bring the extraction into your own workflow.

Clairic Studio is a desktop workbench for [Cloudflare's Clairic model](https://developers.cloudflare.com/workers-ai/models/clairic/), built with Go and Wails.

**Windows x64 · One executable · Five code exporters · [MIT licensed](LICENSE)**

## See the whole workflow

**Document → Fields → Typed results → Working code**

| Bring your text | Shape the answer | Take it with you |
| --- | --- | --- |
| Paste into the editable document pane or open a UTF-8 file. | Add typed fields, describe what you want, and inspect results and confidence. | Copy JSON, save a reusable schema, or export a standalone program. |

- **Iterate while you type.** Auto-preview waits for a one-second pause after an edit. **Extract now** also works with Auto-preview on or off.
- **Inspect the source.** Click a text result to highlight its verified source span. Computed values and predictions may have no direct span.
- **Keep the full answer.** Explore confidence, nulls, ranges, forecasts, and other model metadata alongside the values.
- **Make it your workspace.** Resize the document, field editor, and results panes. Switch between light and dark appearances. Start with an empty document and fields; restore a saved draft explicitly when you want it.
- **Leave with code.** Preview, copy, or save standalone programs that call Cloudflare directly, or PowerShell programs using PSAIClairic.

## Run Studio

**[Download ClairicStudio.exe for Windows x64](https://github.com/dfinke/ClairicStudio/releases/latest/download/ClairicStudio.exe)**, or find the ready-to-run EXE at **`build/bin/ClairicStudio.exe`** after cloning. You can copy that one EXE to another folder or Windows x64 computer; its interface is embedded. Running Studio does not require Go, Node.js, or PowerShell. Windows needs Microsoft Edge WebView2, normally present on current Windows installations. The current executable is unsigned.

Get the Windows executable and its SHA-256 checksum from [GitHub Releases](https://github.com/dfinke/ClairicStudio/releases/latest). The repository also includes the executable alongside its source. Temporary build outputs stay excluded from Git.

1. Connect with your **Cloudflare account ID** and an API token with **Workers AI Read** permission for that account. Studio checks `CLOUDFLARE_ACCOUNT_ID` and `CLOUDFLARE_AUTH_TOKEN` (or the `CLOUDFLARE_API_TOKEN` alias) in its launch environment, Windows user/system environment settings, and its saved credential-store connection. The setup screen appears only when no complete connection is found.
2. Choose **Load example**, paste text, or choose **Open text file**.
3. Add fields and describe what each should contain. Choose **Extract now**, or edit with Auto-preview enabled.
4. Inspect the values, confidence, and source spans. **Copy JSON** copies just the current extracted values.
5. Choose **Export code**, select a language and document-input mode, then **Copy code** or **Save program**. Successful copies briefly show **✓ Copied**.

Schema editing and code export work without a connection. Extraction requires internet access. Model requests send your text and schema to Cloudflare and use your Workers AI allowance. **Test extraction** sends one small request.

**Remember** stores credentials in the operating system credential store. Leaving it unchecked keeps entered credentials for the current session. Complete environment credentials take priority on launch: process/session, Windows user, then Windows system settings. Incomplete settings do not hide a complete saved connection. Restart Studio after changing settings. Tokens are never included in workspace JSON, schema files, or generated programs. **Forget saved credentials** removes the stored connection without changing environment variables.

### Credentials set in a PowerShell profile or session

Windows does not share one PowerShell process's environment with an EXE launched by double-clicking. If your profile sets the variables, open the PowerShell session where they work, then launch Studio from that session:

```powershell
& 'C:\path\to\ClairicStudio.exe'
```

Studio detects the inherited connection. To make future double-click launches work, open **Cloudflare connected**, check **Remember**, and choose **Use credentials**. Leave the token blank to reuse the detected token. Alternatively, set the variables in Windows user/system Environment Variables. Studio does not execute PowerShell profiles or silently persist credentials.

## Your first demo: invoice to automation

Choose **Load example**, or open [invoice.txt](samples/invoice.txt) and [invoice.schema.json](samples/invoice.schema.json).

The fictional document describes four pallets at $250 each, $80 shipping, and net-30 payment terms. Inspect the returned answers against the source:

| Field | Type | Expected value in Load example |
| --- | --- | --- |
| `invoice_number` | text | `INV-2291` |
| `vendor` | text | `Northwind Supply` |
| `total` | number | `1080` USD, including shipping |
| `due_date` | date | `2026-10-02`, calculated from the issue date and terms |
| `po_number` | text | `null`, because no purchase order is present |

The file-based invoice includes purchase order `PO-1042`; use it to contrast a present purchase order with the built-in example's `null`.

Click the vendor result to inspect its source. Refine an instruction and watch the preview status. Copy the JSON, then export PowerShell with **Embed the current document** and run the copied program. Switch to Python to see the same extraction in another language.

For broader exploration, open [all-types.txt](samples/all-types.txt) and [all-types.schema.json](samples/all-types.schema.json). They exercise all nine field types: **text, number, date, time, datetime, duration, list, estimate, and forecast**.

These are model outputs: verify amounts, dates, calculations, and source matches. Confidence is not a guarantee. Estimates and forecasts are predictions rather than facts established by the document.

## Export once, run anywhere your language runs

Exported programs use the selected schema and read credentials from the two environment variables above. Standalone exports call the Cloudflare HTTP API directly; the optional PowerShell module style calls the same API through PSAIClairic. Full answers and metadata remain in the output. Conditional blocks use explicit braces and separate lines where the language supports them; Python uses indented blocks.

| Language | Runtime needed for the exported program |
| --- | --- |
| PowerShell | PowerShell 7.2+; standalone by default, or PSAIClairic 0.2.0+ for module style |
| Python | Python 3; standard library only |
| TypeScript | Node.js 18+ and a TypeScript runner such as `tsx` |
| Go | Go 1.22+; standard library only |
| Rust | A Cargo project with `reqwest` 0.12 (`blocking`, `json`) and `serde_json` 1; setup instructions appear in the exported file |

When PowerShell is selected, **Export style** offers **Standalone HTTP script** (the default) or **Use PSAIClairic module**. Module style builds ordered fields with `New-ClairicField` and runs `Invoke-Clairic`. It preserves the full result and includes a commented `ConvertFrom-ClairicAnswer` pipeline for projecting just the values. Install or import [PSAIClairic](https://github.com/dfinke/PSAIClairic) before running that style; Studio itself does not require the module.

**Copy code** and **Copy JSON** show **✓ Copied** only after a successful clipboard write, reset after two seconds, and clear when the corresponding content changes.

Choose the input mode explicitly:

- **Read from a file when the program runs** keeps the document out of the program and requires a file path. For PowerShell: `./Extract-Document.ps1 -Path ./invoice.txt`.
- **Embed the current document** includes the current text and runs without a path. For PowerShell: `./Extract-Document.ps1`. An optional file argument overrides the embedded document.

Cloudflare credentials saved by Studio are not automatically available to exported programs; those programs use their launch environment.

## Schemas and workspace recovery

Every launch starts with an **empty document and no fields**. Appearance is remembered. Fields, appearance, and opened file paths save automatically to `%LOCALAPPDATA%/ClairicStudio/workspace.json`. Pasted or edited document contents and results stay in memory and are not restored. Editing an opened document does not overwrite the original file.

Choose **Restore draft** (also available from File) to recover the last saved workspace, including an earlier PowerShell designer draft. Recovery is explicit and makes no model request. Choose **Extract now** when ready. Simply opening and closing the empty app leaves the saved draft intact. **New schema** keeps the document and provides Undo until your next edit.

Use **Save schema as** for a portable version 1 JSON schema compatible with the PowerShell ClairicSchemaBuddy format. Draft autosaves are separate from explicitly saved schema files. Explicit draft recovery can migrate earlier PowerShell designer workspaces while leaving the original files in place.

```json
{
  "version": 1,
  "fields": [
    {
      "name": "invoice_number",
      "type": "text",
      "instructions": "Complete invoice number",
      "settings": {}
    },
    {
      "name": "total",
      "type": "number",
      "instructions": "Total amount due including shipping",
      "settings": { "unit": "USD", "decimals": 2 }
    }
  ]
}
```

| Action | Shortcut |
| --- | --- |
| New schema | Ctrl+N |
| Open schema | Ctrl+O |
| Add field | Ctrl+Shift+F |
| Select the current field name | Alt+N |

## Build from source

Development requires **Go 1.23+**, **Node.js 22.12+ or 20.19+**, npm, **Wails CLI 2.10.2**, and WebView2 on Windows. Install the Wails CLI, clone this repository, and build:

```powershell
git clone https://github.com/dfinke/ClairicStudio.git
cd ClairicStudio
go install github.com/wailsapp/wails/v2/cmd/wails@v2.10.2
./Build.ps1
./build/bin/ClairicStudio.exe
```

The build script installs the locked frontend dependencies, runs tests and Go analysis, then builds the EXE. It finds Wails on PATH or in Go's bin directory. `wails dev` starts live development when Wails is on PATH.

To run the checks separately:

```powershell
cd frontend
npm ci
npm test
npm run build
cd ..
go test ./...
go vet ./...
```

Building the frontend first creates the assets embedded by Go. A desktop startup check makes no model requests:

```powershell
Start-Process ./build/bin/ClairicStudio.exe -Wait -ArgumentList '--smoke', '--smoke-report', "$PWD/smoke-report.json"
Get-Content ./smoke-report.json
```

The report contains `uiReady: true` when the interface and native backend initialize successfully. Automated API tests use a local server; they do not assert live model accuracy.

## Project layout

| Location | Purpose |
| --- | --- |
| `clairic/` | UI-independent schema validation, HTTP client, and five code exporters |
| `frontend/src/` | Document editing, field designer, revision tracking, themes, and results |
| `app.go` | Native file dialogs, clipboard, credentials, workspace persistence, and extraction coordination |
| `samples/` | Fictional invoice and all-field-type demos |
| `build/` | Application icon, reproducible icon generator, and platform build resources |

The app currently targets **Windows x64**. macOS and Linux have not been built or verified. Studio accepts **UTF-8 plain text**; PDF and image ingestion are future work. Limits are 8 MiB per document and 500 fields. Extraction has a 120-second timeout, runs one request at a time, supports cancellation, and ignores responses for outdated revisions. Truncation reported by Cloudflare is shown as a warning. Failed requests are not automatically retried.

## License and credits

[MIT](LICENSE) © 2026 Doug Finke. Third-party components retain their own licenses; the generated Wails runtime is covered by its [upstream MIT license](LICENSES/Wails-MIT.txt).

[Cloudflare Clairic documentation](https://developers.cloudflare.com/workers-ai/models/clairic/) · [Wails documentation](https://wails.io/docs/introduction/) · [Changelog](CHANGELOG.md)

Clairic Studio is an independent project and is not an official Cloudflare application.
