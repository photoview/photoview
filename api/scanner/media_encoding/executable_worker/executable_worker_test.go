package executable_worker

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	_ "github.com/photoview/photoview/api/test_utils/flags"
)

const testdataBinPath = "./test_data/mock_bin"

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

func TestInitFfprobePath(t *testing.T) {
	t.Run("PathFail", func(t *testing.T) {
		SetPathWithCurrent(t, "non_exist_path")

		err := SetFfprobePath()
		if err == nil {
			t.Fatalf("InitFfprobePath() returns nil, want an error")
		}
	})

	t.Run("VersionFail", func(t *testing.T) {
		SetPathWithCurrent(t, testdataBinPath)
		t.Setenv("FAIL_WITH", "expect failure")

		err := SetFfprobePath()
		if err == nil {
			t.Fatalf("InitFfprobePath() returns nil, want an error")
		}
	})

	t.Run("Succeed", func(t *testing.T) {
		SetPathWithCurrent(t, testdataBinPath)

		err := SetFfprobePath()
		if err != nil {
			t.Fatalf("InitFfprobePath() returns %v, want nil", err)
		}
	})
}
