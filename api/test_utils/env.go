package test_utils

import (
	"path/filepath"
	"runtime"
)

// PathFromAPIRoot returns the real path in the API project root.
func PathFromAPIRoot(rootRelatedPath string) string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		panic("Can't get the path of current function. It should not happen.")
	}

	base := filepath.Dir(file)

	return filepath.Join(base, "..", rootRelatedPath)
}
