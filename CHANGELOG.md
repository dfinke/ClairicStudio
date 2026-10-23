# Changelog

## Unreleased

- Launch now starts with an empty document and fields. Saved drafts and earlier designer workspaces are recovered only with Restore draft; unopened drafts stay intact.
- Credential discovery checks the launch environment, Windows user/system variables, the API-token alias, and saved credentials. Partial environment settings no longer hide a saved connection.
- Setup explains PowerShell-session credentials and how Remember enables future double-click launches without retyping a detected token.

- The ready-to-run Windows x64 executable is included in the repository with a README download link.

- Copy code and Copy JSON briefly confirm successful clipboard writes with ✓ Copied; content changes reset the confirmation.
- PowerShell export now offers standalone HTTP and PSAIClairic module styles, with matching schemas, input modes, and request options. Standalone remains the default.

## 0.1.0 — 2026-10-23

Initial desktop version of Clairic Studio.

- Editable text documents, nine Clairic field types, automatic previews, and an Extract now button that works with Auto-preview on or off.
- Typed result tables, JSON copying, answer metadata, exact source-span highlighting, and truncation warnings.
- Standalone PowerShell, Python, TypeScript, Go, and Rust exports with file or embedded document input, code preview, clipboard copying, and program saving.
- Version 1 schema import/export, draft recovery, earlier PowerShell designer workspace migration, and Undo for New schema.
- Resizable panes, light and dark appearances, keyboard shortcuts, and a custom Windows application icon.
- Environment credential discovery and optional operating system credential storage.
- Cancellation, outdated-response handling, local API tests, frontend checks, and an isolated desktop startup check.
- Fictional invoice and all-field-type demos; MIT license.
