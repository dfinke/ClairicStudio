package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/dfinke/ClairicStudio/clairic"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/zalando/go-keyring"
)

type credentials struct {
	Account string `json:"account"`
	Token   string `json:"token"`
	Source  string `json:"-"`
}
type Connection struct {
	Account    string `json:"account"`
	Configured bool   `json:"configured"`
	Source     string `json:"source"`
	Message    string `json:"message"`
}
type Workspace struct {
	Schema       clairic.Schema `json:"schema"`
	Theme        string         `json:"theme"`
	DocumentPath string         `json:"document_path"`
	SchemaPath   string         `json:"schema_path"`
}
type Bootstrap struct {
	Workspace  Workspace  `json:"workspace"`
	Document   string     `json:"document"`
	Connection Connection `json:"connection"`
	Notice     string     `json:"notice"`
}
type Document struct {
	Path string `json:"path"`
	Text string `json:"text"`
}
type SchemaFile struct {
	Path   string         `json:"path"`
	Schema clairic.Schema `json:"schema"`
}
type App struct {
	ctx          context.Context
	mu           sync.Mutex
	workspace    Workspace
	creds        credentials
	dir          string
	client       *clairic.Client
	active       bool
	cancel       context.CancelFunc
	saveTimer    *time.Timer
	smoke        bool
	report       string
	ready        bool
	startupError string
}

func NewApp() *App {
	return &App{client: clairic.NewClient(), workspace: Workspace{Schema: clairic.Schema{Version: 1, Fields: []clairic.Field{}}, Theme: "dark"}}
}
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	if a.dir == "" {
		dir, err := os.UserConfigDir()
		if err != nil {
			a.startupError = "Could not locate your settings directory"
		} else {
			if local := os.Getenv("LOCALAPPDATA"); local != "" {
				dir = local
			}
			a.dir = filepath.Join(dir, "ClairicStudio")
		}
	}
	current := filepath.Join(a.dir, "workspace.json")
	data, err := os.ReadFile(current)
	migrated := false
	if os.IsNotExist(err) && !a.smoke {
		for _, old := range []string{"ClairicSchemaBuddy", "PSClairicSchemaDesigner", filepath.Join("PSAIClairic", "SchemaBuddy")} {
			legacy := filepath.Join(os.Getenv("LOCALAPPDATA"), old, "workspace.json")
			if b, e := os.ReadFile(legacy); e == nil {
				data = b
				err = nil
				migrated = true
				break
			}
		}
	}
	if err == nil {
		var saved Workspace
		if len(data) > 2<<20 || json.Unmarshal(data, &saved) != nil || saved.Schema.Version != 1 {
			a.startupError = "Saved workspace could not be read. Open a schema or start fresh."
		} else {
			raw, _ := json.Marshal(saved.Schema)
			if s, e := clairic.ReadSchema(raw); e == nil {
				saved.Schema = s
				a.workspace = saved
			} else {
				a.startupError = "Saved schema could not be read. Open a schema or start fresh."
			}
		}
	} else if !os.IsNotExist(err) {
		a.startupError = "Workspace could not be read; changes can still be exported."
	}
	a.workspace.Theme = strings.ToLower(a.workspace.Theme)
	if a.workspace.Theme != "light" {
		a.workspace.Theme = "dark"
	}
	account, token := os.Getenv("CLOUDFLARE_ACCOUNT_ID"), os.Getenv("CLOUDFLARE_AUTH_TOKEN")
	if account != "" || token != "" {
		a.creds = credentials{account, token, "environment"}
	} else if !a.smoke {
		if secret, e := keyring.Get("ClairicStudio", "cloudflare"); e == nil {
			_ = json.Unmarshal([]byte(secret), &a.creds)
			a.creds.Source = "credential store"
		}
	}
	if migrated {
		_ = a.writeWorkspaceLocked()
	}
	if a.smoke {
		go func() {
			time.Sleep(15 * time.Second)
			a.mu.Lock()
			ready := a.ready
			a.mu.Unlock()
			if !ready {
				a.finishSmoke("Frontend startup timed out")
			}
		}()
	}
}
func (a *App) beforeClose(ctx context.Context) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.saveTimer != nil {
		a.saveTimer.Stop()
	}
	_ = a.writeWorkspaceLocked()
	if a.cancel != nil {
		a.cancel()
	}
	return false
}
func (a *App) Ready() {
	a.mu.Lock()
	a.ready = true
	a.mu.Unlock()
	if a.smoke {
		a.finishSmoke("")
	}
}
func (a *App) finishSmoke(problem string) {
	data, _ := json.Marshal(map[string]any{"uiReady": problem == "", "error": problem, "application": "Clairic Studio"})
	if a.report != "" {
		_ = os.WriteFile(a.report, data, 0600)
	}
	runtime.Quit(a.ctx)
}
func (a *App) Bootstrap() Bootstrap {
	a.mu.Lock()
	defer a.mu.Unlock()
	b := Bootstrap{Workspace: a.workspace, Connection: a.connectionLocked(), Notice: a.startupError}
	if a.workspace.DocumentPath != "" {
		if text, err := readText(a.workspace.DocumentPath); err == nil {
			b.Document = text
		} else {
			b.Notice = "Workspace restored; the previous document is unavailable. Paste text or open it again."
		}
	}
	if b.Notice == "" {
		b.Notice = "Workspace restored without making a model request."
	}
	return b
}
func (a *App) connectionLocked() Connection {
	ok := clairic.AccountPattern.MatchString(a.creds.Account) && strings.TrimSpace(a.creds.Token) != ""
	msg := "Connect Cloudflare to extract. You can design and export schemas now."
	if ok {
		msg = "Cloudflare credentials available (" + a.creds.Source + ")."
	}
	return Connection{a.creds.Account, ok, a.creds.Source, msg}
}
func (a *App) ConfigureConnection(account, token string, remember bool) (Connection, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	account = strings.TrimSpace(account)
	token = strings.TrimSpace(token)
	if token == "" && account == a.creds.Account {
		token = a.creds.Token
	}
	if !clairic.AccountPattern.MatchString(account) || token == "" {
		return a.connectionLocked(), errors.New("Enter the 32-character account ID and an API token")
	}
	next := credentials{account, token, "this session"}
	if remember {
		raw, _ := json.Marshal(next)
		if err := keyring.Set("ClairicStudio", "cloudflare", string(raw)); err != nil {
			return a.connectionLocked(), errors.New("Could not save to the OS credential store. Uncheck Remember to use this session only")
		}
		next.Source = "credential store"
	}
	a.creds = next
	return a.connectionLocked(), nil
}
func (a *App) ForgetConnection() (Connection, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if err := keyring.Delete("ClairicStudio", "cloudflare"); err != nil && err != keyring.ErrNotFound {
		return a.connectionLocked(), errors.New("Could not remove saved credentials from the OS store")
	}
	a.creds = credentials{os.Getenv("CLOUDFLARE_ACCOUNT_ID"), os.Getenv("CLOUDFLARE_AUTH_TOKEN"), "environment"}
	return a.connectionLocked(), nil
}
func (a *App) TestConnection() (string, error) {
	s := clairic.Schema{Version: 1, Fields: []clairic.Field{{Name: "code", Type: "text", Instructions: "Extract the connection code", Settings: map[string]any{}}}}
	_, err := a.Extract("Connection code STUDIO-OK.", s)
	if err != nil {
		return "", err
	}
	return "Cloudflare accepted a small extraction request.", nil
}
func (a *App) Validate(state string, s clairic.Schema) string {
	_, err := clairic.BuildRequest(state, s)
	if err != nil {
		return err.Error()
	}
	return ""
}
func (a *App) Extract(state string, s clairic.Schema) (clairic.Result, error) {
	request, err := clairic.BuildRequest(state, s)
	if err != nil {
		return clairic.Result{}, err
	}
	a.mu.Lock()
	if a.active {
		a.mu.Unlock()
		return clairic.Result{}, errors.New("An extraction is already running")
	}
	if !a.connectionLocked().Configured {
		a.mu.Unlock()
		return clairic.Result{}, errors.New("Connect your Cloudflare account first")
	}
	creds := a.creds
	a.active = true
	ctx, cancel := context.WithCancel(a.ctx)
	a.cancel = cancel
	a.mu.Unlock()
	defer func() { cancel(); a.mu.Lock(); a.active = false; a.cancel = nil; a.mu.Unlock() }()
	return a.client.Extract(ctx, creds.Account, creds.Token, request)
}
func (a *App) CancelExtraction() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.cancel != nil {
		a.cancel()
	}
}
func (a *App) GenerateCode(language, state string, s clairic.Schema, embed bool, style string) (clairic.CodeExport, error) {
	return clairic.ExportProgramWithStyle(language, state, s, embed, style)
}
func (a *App) SaveCode(language, state string, s clairic.Schema, embed bool, style string) (string, error) {
	e, err := a.GenerateCode(language, state, s, embed, style)
	if err != nil {
		return "", err
	}
	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{Title: "Save " + language + " program", DefaultFilename: e.Filename})
	if err != nil || path == "" {
		return "", err
	}
	if err = os.WriteFile(path, []byte(e.Code), 0600); err != nil {
		return "", errors.New("Could not write the exported program")
	}
	return path, nil
}
func readText(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil || info.IsDir() || info.Size() > 8<<20 {
		return "", errors.New("Choose a plain-text file no larger than 8 MiB")
	}
	data, err := os.ReadFile(path)
	if err != nil || !utf8.Valid(data) || strings.ContainsRune(string(data), 0) {
		return "", errors.New("Choose a UTF-8 text file")
	}
	return strings.TrimPrefix(string(data), "\ufeff"), nil
}
func (a *App) OpenDocument() (Document, error) {
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{Title: "Open document", Filters: []runtime.FileFilter{{DisplayName: "Text documents", Pattern: "*.txt;*.md;*.csv;*.json;*.log;*.eml"}, {DisplayName: "All files", Pattern: "*"}}})
	if err != nil || path == "" {
		return Document{}, err
	}
	text, err := readText(path)
	return Document{path, text}, err
}
func (a *App) OpenSchema() (SchemaFile, error) {
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{Title: "Open extraction schema", Filters: []runtime.FileFilter{{DisplayName: "Schema JSON", Pattern: "*.json"}}})
	if err != nil || path == "" {
		return SchemaFile{}, err
	}
	info, err := os.Stat(path)
	if err != nil || info.Size() > 2<<20 {
		return SchemaFile{}, errors.New("Choose a schema no larger than 2 MiB")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return SchemaFile{}, errors.New("Could not read schema")
	}
	s, err := clairic.ReadSchema(data)
	return SchemaFile{path, s}, err
}
func (a *App) SaveSchema(s clairic.Schema) (string, error) {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return "", errors.New("Schema could not be encoded")
	}
	if _, err = clairic.ReadSchema(data); err != nil {
		return "", err
	}
	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{Title: "Save extraction schema", DefaultFilename: "extraction.schema.json", Filters: []runtime.FileFilter{{DisplayName: "Schema JSON", Pattern: "*.json"}}})
	if err != nil || path == "" {
		return "", err
	}
	if err = atomicWrite(path, data); err != nil {
		return "", errors.New("Could not save schema")
	}
	return path, nil
}
func (a *App) UpdateWorkspace(w Workspace) error {
	data, err := json.Marshal(w.Schema)
	if err != nil {
		return errors.New("Workspace could not be encoded")
	}
	if _, err = clairic.ReadSchema(data); err != nil {
		return err
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	a.workspace = w
	if a.saveTimer != nil {
		a.saveTimer.Stop()
	}
	a.saveTimer = time.AfterFunc(450*time.Millisecond, func() {
		a.mu.Lock()
		defer a.mu.Unlock()
		if err := a.writeWorkspaceLocked(); err != nil {
			runtime.EventsEmit(a.ctx, "studio:save-error", "Workspace could not be saved; export your schema to retain it.")
		}
	})
	return nil
}
func (a *App) writeWorkspaceLocked() error {
	data, err := json.MarshalIndent(a.workspace, "", "  ")
	if err != nil {
		return err
	}
	return atomicWrite(filepath.Join(a.dir, "workspace.json"), data)
}
func atomicWrite(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(path), ".studio-*.tmp")
	if err != nil {
		return err
	}
	tmp := f.Name()
	defer os.Remove(tmp)
	if _, err = f.Write(data); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	// Windows cannot atomically rename over an existing file using os.Rename.
	return replaceFile(tmp, path)
}
func (a *App) CopyText(text string) error {
	if err := runtime.ClipboardSetText(a.ctx, text); err != nil {
		return fmt.Errorf("Clipboard could not be updated")
	}
	return nil
}
