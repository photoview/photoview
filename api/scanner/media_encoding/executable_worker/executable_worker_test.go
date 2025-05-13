package executable_worker

import (
	"testing"

	"github.com/photoview/photoview/api/test_utils"
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
		test_utils.SetPathWithCurrent(t, testdataBinPath)
		t.Setenv("FAIL_WITH", "expect failure")

		err := SetFfprobePath()
		if err == nil {
			t.Fatalf("InitFfprobePath() returns nil, want an error")
		}
	})

	t.Run("Succeed", func(t *testing.T) {
		test_utils.SetPathWithCurrent(t, testdataBinPath)

		err := SetFfprobePath()
		if err != nil {
			t.Fatalf("InitFfprobePath() returns %v, want nil", err)
		}
	})
}
