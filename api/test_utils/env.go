package test_utils

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// SetPathWithCurrent sets PATH env to `paths` in the directory of testing files. The PATH will restore to the previous value when the test is done.
func SetPathWithCurrent(t *testing.T, paths ...string) {
	_, file, _, ok := runtime.Caller(1)
	if !ok {
		t.Log("Can't get the test file. Ignore setting PATH.")
		return
	}

	base := filepath.Dir(file)

	for i, path := range paths {
		paths[i] = filepath.Join(base, path)
	}

	t.Setenv("PATH", strings.Join(paths, ":"))
}

// PathFromAPIRoot returns the real path in the API project root.
func PathFromAPIRoot(rootRelatedPath string) string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		panic("Can't get the path of current function. It should not happen.")
	}

	base := filepath.Dir(file)

	return filepath.Join(base, "..", rootRelatedPath)
}
