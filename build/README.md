# Application build resources

`appicon.png` and `windows/icon.ico` are the custom Clairic Studio icons. Regenerate both on Windows with PowerShell 7:

```powershell
./build/Generate-Icon.ps1
```

The generator renders its vector geometry at 16, 20, 24, 32, 40, 48, 64, 128, and 256 pixels and packs the PNG frames into the Windows icon. Rebuild the app after changing icon resources.

- `windows/` contains the executable metadata, application manifest, and Wails installer templates. An installer is not required to run the built EXE.
- `darwin/` contains Wails macOS metadata templates. macOS builds are not currently verified.
- `bin/ClairicStudio.exe` is the tracked Windows executable. Other local build outputs in `bin/` stay ignored.

Use [Build.ps1](../Build.ps1) from the repository root to run checks and build the Windows desktop app. See the [main README](../README.md) for prerequisites and usage.
