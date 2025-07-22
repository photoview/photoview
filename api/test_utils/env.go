package test_utils

import (
	"path/filepath"
	"runtime"
)

// PathFromAPIRoot returns the real path in the API project root.
func PathFromAPIRoot(rootRelatedPaths ...string) string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		panic("Can't get the path of current function. It should not happen.")
	}

	base := filepath.Dir(file)
	args := make([]string, 0, len(rootRelatedPaths)+2)
	args = append(args, []string{base, ".."}...)
	args = append(args, rootRelatedPaths...)

	return filepath.Join(args...)
}
