package main

import (
	"embed"
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	smoke := flag.Bool("smoke", false, "Verify desktop startup without model calls")
	report := flag.String("smoke-report", "", "File for the desktop smoke result")
	flag.Parse()
	app := NewApp()
	app.smoke = *smoke
	app.report = *report
	if *smoke {
		dir, err := os.MkdirTemp("", "ClairicStudio-smoke-")
		if err != nil {
			log.Fatal(err)
		}
		defer os.RemoveAll(dir)
		app.dir = dir
	}
	userData, _ := os.UserCacheDir()
	if *smoke {
		userData = app.dir
	}
	userData = filepath.Join(userData, "ClairicStudio", "WebView2")
	appMenu := menu.NewMenu()
	file := appMenu.AddSubmenu("File")
	emit := func(action string) func(*menu.CallbackData) {
		return func(_ *menu.CallbackData) { runtime.EventsEmit(app.ctx, "studio:action", action) }
	}
	file.AddText("New schema", nil, emit("new"))
	file.AddText("Open schema…", nil, emit("open-schema"))
	file.AddText("Save schema as…", nil, emit("save-schema"))
	file.AddSeparator()
	file.AddText("Open text file…", nil, emit("open-document"))
	file.AddText("Export code…", nil, emit("export"))
	file.AddSeparator()
	file.AddText("Quit", nil, func(_ *menu.CallbackData) { runtime.Quit(app.ctx) })
	edit := appMenu.AddSubmenu("Edit")
	edit.AddText("Add field", nil, emit("add-field"))
	edit.AddText("Undo new schema", nil, emit("undo"))
	settings := appMenu.AddSubmenu("Settings")
	settings.AddText("Cloudflare connection…", nil, emit("connect"))
	settings.AddText("Toggle appearance", nil, emit("theme"))
	err := wails.Run(&options.App{Title: "Clairic Studio", Width: 1440, Height: 820, MinWidth: 1000, MinHeight: 700, StartHidden: *smoke, Menu: appMenu,
		Windows:          &windows.Options{WebviewUserDataPath: userData},
		BackgroundColour: &options.RGBA{R: 16, G: 23, B: 34, A: 255}, AssetServer: &assetserver.Options{Assets: assets},
		OnStartup: app.startup, OnBeforeClose: app.beforeClose, Bind: []interface{}{app}, EnableDefaultContextMenu: true})
	if err != nil {
		if *report != "" {
			_ = os.WriteFile(*report, []byte(`{"uiReady":false,"error":"Wails startup failed"}`), 0600)
		}
		log.Fatal(err)
	}
}
