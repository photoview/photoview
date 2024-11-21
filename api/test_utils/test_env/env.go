package test_env

import (
	"flag"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type integration_options struct {
	Database   *bool
	Filesystem *bool
}

// Fake flags to be compatible with current test command.
var integration_flags integration_options = integration_options{
	Database:   flag.Bool("database", false, "run database integration tests"),
	Filesystem: flag.Bool("filesystem", false, "run filesystem integration tests"),
}

func SetPathWithCurrent(paths ...string) func() {
	_, file, _, ok := runtime.Caller(1)
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

func PathFromAPIRoot(rootRelatedPath string) string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		panic("Can't get the path of current function. It should not happen.")
	}

	base := filepath.Dir(file)

	return filepath.Join(base, "../..", rootRelatedPath)
}
