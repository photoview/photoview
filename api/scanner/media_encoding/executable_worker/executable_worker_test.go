package executable_worker_test

import (
	"os"
	"testing"

	"github.com/photoview/photoview/api/scanner/media_encoding/executable_worker"
	"github.com/photoview/photoview/api/test_utils"
)

func TestMain(m *testing.M) {
	os.Exit(test_utils.IntegrationTestRun(m))
}

const testdataBinPath = "./testdata/bin"

func setPathWithTestdataBin() func() {
	return test_utils.SetPathWithCurrent(testdataBinPath)
}

func TestInitFfprobePath(t *testing.T) {
	t.Run("PathFail", func(t *testing.T) {
		err := executable_worker.SetFfprobePath()
		if err == nil {
			t.Fatalf("InitFfprobePath() returns nil, want an error")
		}
	})

	t.Run("VersionFail", func(t *testing.T) {
		donePath := setPathWithTestdataBin()
		defer donePath()

		doneEnv := test_utils.SetEnv("FAIL_WITH", "expect failure")
		defer doneEnv()

		err := executable_worker.SetFfprobePath()
		if err == nil {
			t.Fatalf("InitFfprobePath() returns nil, want an error")
		}
	})

	t.Run("Succeed", func(t *testing.T) {
		donePath := setPathWithTestdataBin()
		defer donePath()

		err := executable_worker.SetFfprobePath()
		if err != nil {
			t.Fatalf("InitFfprobePath() returns %v, want nil", err)
		}
	})
}
