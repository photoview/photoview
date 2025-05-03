package executable_worker

import (
	"testing"

	"github.com/kkovaletp/photoview/api/test_utils"
)

const testdataBinPath = "./test_data/mock_bin"

func TestInitFfprobePath(t *testing.T) {
	t.Run("PathFail", func(t *testing.T) {
		err := SetFfprobePath()
		if err == nil {
			t.Fatalf("InitFfprobePath() returns nil, want an error")
		}
	})

	t.Run("VersionFail", func(t *testing.T) {
		donePath := test_utils.SetPathWithCurrent(testdataBinPath)
		defer donePath()

		doneEnv := test_utils.SetEnv("FAIL_WITH", "expect failure")
		defer doneEnv()

		err := SetFfprobePath()
		if err == nil {
			t.Fatalf("InitFfprobePath() returns nil, want an error")
		}
	})

	t.Run("Succeed", func(t *testing.T) {
		donePath := test_utils.SetPathWithCurrent(testdataBinPath)
		defer donePath()

		err := SetFfprobePath()
		if err != nil {
			t.Fatalf("InitFfprobePath() returns %v, want nil", err)
		}
	})
}
