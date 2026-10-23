# Changelog

## Unreleased

## 0.1.0 — 2026-10-23

Initial public desktop release of Clairic Studio.

- Editable text documents, nine Clairic field types, automatic previews, and an Extract now button that works with Auto-preview on or off.
- Typed result tables, JSON copying, answer metadata, exact source-span highlighting, and truncation warnings.
- Standalone PowerShell, Python, TypeScript, Go, and Rust exports with file or embedded document input, code preview, clipboard copying, and program saving.
- Optional PSAIClairic module style for PowerShell exports, preserving schemas, input modes, and request options.
- Copy code and Copy JSON briefly confirm successful clipboard writes with ✓ Copied; content changes reset the confirmation.
- Empty document and fields on launch. Restore draft explicitly recovers saved workspaces and earlier PowerShell designer drafts; unopened drafts stay intact.
- Version 1 schema import/export, automatic draft saving, and Undo for New schema.
- Resizable panes, light and dark appearances, keyboard shortcuts, and a custom Windows application icon.
- Credential discovery checks the launch environment, Windows user/system variables, the API-token alias, and saved credentials. Partial settings do not hide a saved connection.
- Optional operating system credential storage, with setup instructions for PowerShell-session credentials and future double-click launches.
- Cancellation, outdated-response handling, local API tests, frontend checks, and an isolated desktop startup check.
- Ready-to-run Windows x64 executable and SHA-256 checksum available from GitHub Releases; the executable is also included in the repository.
- Fictional invoice and all-field-type demos; MIT license.
