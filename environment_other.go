//go:build !windows

package main

func persistedEnvironment(scope, name string) string {
	return ""
}
