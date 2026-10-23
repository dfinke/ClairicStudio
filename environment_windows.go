//go:build windows

package main

import "golang.org/x/sys/windows/registry"

func persistedEnvironment(scope, name string) string {
	root, path := registry.CURRENT_USER, `Environment`
	if scope == "machine" {
		root, path = registry.LOCAL_MACHINE, `SYSTEM\CurrentControlSet\Control\Session Manager\Environment`
	} else if scope != "user" {
		return ""
	}
	key, err := registry.OpenKey(root, path, registry.QUERY_VALUE)
	if err != nil {
		return ""
	}
	defer key.Close()
	value, kind, err := key.GetStringValue(name)
	if err != nil {
		return ""
	}
	if kind == registry.EXPAND_SZ {
		if expanded, err := registry.ExpandString(value); err == nil {
			return expanded
		}
	}
	return value
}
