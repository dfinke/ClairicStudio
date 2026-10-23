package main

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/dfinke/ClairicStudio/clairic"
	"github.com/zalando/go-keyring"
)

type environmentLookup func(scope, name string) string

func credentialEnvironment(scope, name string) string {
	if scope == "process" {
		return os.Getenv(name)
	}
	return persistedEnvironment(scope, name)
}

func readStoredCredentials() credentials {
	var saved credentials
	if secret, err := keyring.Get("ClairicStudio", "cloudflare"); err == nil {
		if json.Unmarshal([]byte(secret), &saved) == nil {
			saved.Source = "credential store"
		}
	}
	return saved
}

func completeCredentials(c credentials) bool {
	return clairic.AccountPattern.MatchString(c.Account) && c.Token != ""
}

// Check every supported scope before showing setup. Incomplete environment
// settings must not hide a complete connection in the OS credential store.
func discoverCredentials(lookup environmentLookup, loadStored func() credentials) credentials {
	var found credentials
	var accountScope, tokenScope string
	for _, scope := range []string{"process", "user", "machine"} {
		account := strings.TrimSpace(lookup(scope, "CLOUDFLARE_ACCOUNT_ID"))
		if found.Account == "" && clairic.AccountPattern.MatchString(account) {
			found.Account, accountScope = account, scope
		}
		if found.Token == "" {
			for _, name := range []string{"CLOUDFLARE_AUTH_TOKEN", "CLOUDFLARE_API_TOKEN"} {
				if token := strings.TrimSpace(lookup(scope, name)); token != "" {
					found.Token, tokenScope = token, scope
					break
				}
			}
		}
	}
	if completeCredentials(found) {
		found.Source = "environment (" + accountScope
		if tokenScope != accountScope {
			found.Source += "/" + tokenScope
		}
		found.Source += ")"
		return found
	}
	if loadStored != nil {
		saved := loadStored()
		saved.Account, saved.Token = strings.TrimSpace(saved.Account), strings.TrimSpace(saved.Token)
		if completeCredentials(saved) {
			saved.Source = "credential store"
			return saved
		}
	}
	// An incomplete connection may prefill the account, but a stray token is not used.
	return credentials{Account: found.Account}
}
