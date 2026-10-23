package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dfinke/ClairicStudio/clairic"
)

func lookupFrom(values map[string]string) environmentLookup {
	return func(scope, name string) string { return values[scope+"/"+name] }
}

func TestCredentialDiscoveryScopesAndAliases(t *testing.T) {
	account := strings.Repeat("a", 32)
	for _, scope := range []string{"process", "user", "machine"} {
		for _, tokenName := range []string{"CLOUDFLARE_AUTH_TOKEN", "CLOUDFLARE_API_TOKEN"} {
			t.Run(scope+"/"+tokenName, func(t *testing.T) {
				lookup := lookupFrom(map[string]string{scope + "/CLOUDFLARE_ACCOUNT_ID": " " + account + " ", scope + "/" + tokenName: " token "})
				got := discoverCredentials(lookup, func() credentials { t.Fatal("valid environment must take precedence"); return credentials{} })
				if got.Account != account || got.Token != "token" || got.Source != "environment ("+scope+")" {
					t.Fatalf("credential discovery failed for %s", scope)
				}
			})
		}
	}
}

func TestCredentialScopePriorityAndMixedScopes(t *testing.T) {
	account := strings.Repeat("b", 32)
	lookup := lookupFrom(map[string]string{
		"process/CLOUDFLARE_API_TOKEN": "process-token", "user/CLOUDFLARE_ACCOUNT_ID": account,
		"user/CLOUDFLARE_AUTH_TOKEN": "user-token", "machine/CLOUDFLARE_ACCOUNT_ID": strings.Repeat("c", 32),
	})
	got := discoverCredentials(lookup, nil)
	if got.Account != account || got.Token != "process-token" || got.Source != "environment (user/process)" {
		t.Fatal("scope precedence did not match the launch environment")
	}
}

func TestPartialEnvironmentDoesNotHideSavedConnection(t *testing.T) {
	for _, value := range []string{"", "not-an-account", strings.Repeat("d", 32)} {
		lookup := lookupFrom(map[string]string{"process/CLOUDFLARE_ACCOUNT_ID": value})
		got := discoverCredentials(lookup, func() credentials { return credentials{Account: strings.Repeat("e", 32), Token: "saved-token"} })
		if !completeCredentials(got) || got.Source != "credential store" {
			t.Fatal("partial environment prevented discovery of a saved connection")
		}
	}
	if completeCredentials(discoverCredentials(lookupFrom(map[string]string{"process/CLOUDFLARE_API_TOKEN": "stray-token"}), nil)) {
		t.Fatal("token without an account should require setup")
	}
}

func TestConfiguredStartupIsEmptyAndDoesNotReadDocument(t *testing.T) {
	app := NewApp()
	app.dir = t.TempDir()
	app.environment = lookupFrom(map[string]string{"user/CLOUDFLARE_ACCOUNT_ID": strings.Repeat("a", 32), "user/CLOUDFLARE_API_TOKEN": "PRIVATE-STARTUP-TOKEN"})
	app.storedCredentials = func() credentials {
		t.Fatal("configured Windows settings must not read the credential store")
		return credentials{}
	}
	draft := Workspace{Schema: clairic.Schema{Version: 1, Fields: []clairic.Field{{Name: "vendor", Type: "text"}}}, Theme: "Light", DocumentPath: filepath.Join(app.dir, "missing-document.txt"), SchemaPath: "old.schema.json"}
	data, _ := json.Marshal(draft)
	path := filepath.Join(app.dir, "workspace.json")
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	app.startup(context.Background())
	boot := app.Bootstrap()
	if !boot.Connection.Configured || boot.Document != "" || len(boot.Workspace.Schema.Fields) != 0 || boot.Workspace.DocumentPath != "" || boot.Workspace.SchemaPath != "" || boot.Workspace.Theme != "light" || !boot.HasSavedDraft || app.active {
		t.Fatal("configured startup must stay blank without reading old documents or extracting")
	}
	raw, _ := json.Marshal(boot)
	if strings.Contains(string(raw), "PRIVATE-STARTUP-TOKEN") {
		t.Fatal("token was exposed to the frontend")
	}
	app.beforeClose(context.Background())
	after, _ := os.ReadFile(path)
	if string(after) != string(data) {
		t.Fatal("opening and closing an empty window changed the saved draft")
	}
	app.workspace.Theme = "dark"
	if err := app.writeWorkspaceLocked(); err != nil {
		t.Fatal(err)
	}
	var saved Workspace
	after, _ = os.ReadFile(path)
	if json.Unmarshal(after, &saved) != nil || len(saved.Schema.Fields) != 1 || saved.Theme != "dark" {
		t.Fatal("appearance changes erased the saved draft")
	}
}

func TestDraftRestoreKeepsCurrentAppearance(t *testing.T) {
	app := NewApp()
	app.dir = t.TempDir()
	app.workspace.Theme = "light"
	app.savedWorkspace = &Workspace{Schema: clairic.Schema{Version: 1, Fields: []clairic.Field{{Name: "vendor", Type: "text"}}}, Theme: "dark"}
	boot, err := app.RestoreWorkspace()
	if err != nil || boot.Workspace.Theme != "light" || len(boot.Workspace.Schema.Fields) != 1 || app.active {
		t.Fatal("draft restore changed the appearance or started extraction")
	}
}
