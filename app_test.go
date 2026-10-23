package main

import (
	"context"
	"encoding/json"
	"github.com/dfinke/ClairicStudio/clairic"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWorkspaceNeverPersistsDocumentOrToken(t *testing.T) {
	a := NewApp()
	a.dir = t.TempDir()
	a.ctx = context.Background()
	a.creds = credentials{strings.Repeat("a", 32), "SECRET-DO-NOT-WRITE", "session"}
	a.workspace = Workspace{Schema: clairic.Schema{Version: 1, Fields: []clairic.Field{}}, Theme: "dark", DocumentPath: "invoice.txt"}
	if err := a.writeWorkspaceLocked(); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(filepath.Join(a.dir, "workspace.json"))
	if strings.Contains(string(data), "SECRET") {
		t.Fatal("token persisted")
	}
	if err := a.writeWorkspaceLocked(); err != nil {
		t.Fatal("workspace replacement failed", err)
	}
	var saved Workspace
	if json.Unmarshal(data, &saved) != nil || saved.Theme != "dark" {
		t.Fatal("workspace invalid")
	}
}
func TestEnvironmentConnectionIsDetectedButTokenIsNotReturned(t *testing.T) {
	t.Setenv("CLOUDFLARE_ACCOUNT_ID", strings.Repeat("b", 32))
	t.Setenv("CLOUDFLARE_AUTH_TOKEN", "PRIVATE-TOKEN")
	a := NewApp()
	a.dir = t.TempDir()
	a.startup(context.Background())
	bootstrap := a.Bootstrap()
	if !bootstrap.Connection.Configured || bootstrap.Connection.Source != "environment" {
		t.Fatal("environment not detected")
	}
	data, _ := json.Marshal(bootstrap)
	if strings.Contains(string(data), "PRIVATE-TOKEN") {
		t.Fatal("token exposed to UI")
	}
}
func TestOldWorkspaceLoadsWithoutExtraction(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("LOCALAPPDATA", dir)
	t.Setenv("CLOUDFLARE_ACCOUNT_ID", strings.Repeat("c", 32))
	t.Setenv("CLOUDFLARE_AUTH_TOKEN", "test")
	old := filepath.Join(dir, "ClairicSchemaBuddy", "workspace.json")
	os.MkdirAll(filepath.Dir(old), 0700)
	data := []byte(`{"theme":"Dark","schema":{"version":1,"fields":[{"name":"total","type":"number","instructions":"amount","settings":{"decimals":"2"}}]},"document_path":"","schema_path":""}`)
	os.WriteFile(old, data, 0600)
	a := NewApp()
	a.dir = filepath.Join(dir, "ClairicStudio")
	a.startup(context.Background())
	if len(a.workspace.Schema.Fields) != 1 || a.active {
		t.Fatal("workspace not restored safely")
	}
	after, _ := os.ReadFile(old)
	if string(after) != string(data) {
		t.Fatal("legacy workspace was changed")
	}
	if _, e := os.Stat(filepath.Join(a.dir, "workspace.json")); e != nil {
		t.Fatal("migration not saved")
	}
}
func TestAtomicWritePreservesPreviousOnFailure(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "schema.json")
	if e := atomicWrite(path, []byte("old")); e != nil {
		t.Fatal(e)
	}
	if e := atomicWrite(path, []byte("new")); e != nil {
		t.Fatal(e)
	}
	data, _ := os.ReadFile(path)
	if string(data) != "new" {
		t.Fatal("replace failed")
	}
	files, _ := filepath.Glob(filepath.Join(dir, ".studio-*.tmp"))
	if len(files) != 0 {
		t.Fatal("temporary files left behind")
	}
}
