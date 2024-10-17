package test_utils

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func SetPathWithCurrent(paths ...string) func() {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return func() {
			// Return an empty function in case of error
		}
	}

	base := filepath.Dir(file)

	for i, path := range paths {
		paths[i] = filepath.Join(base, path)
	}

	originalPath := os.Getenv("PATH")
	os.Setenv("PATH", strings.Join(paths, ":"))

	return func() {
		os.Setenv("PATH", originalPath)
	}
}

func SetEnv(key, value string) func() {
	org := os.Getenv(key)
	os.Setenv(key, value)
	return func() {
		os.Setenv(key, org)
	}
}
