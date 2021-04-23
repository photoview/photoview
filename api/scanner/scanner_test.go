package scanner_test

import (
	"os"
	"testing"

	"github.com/photoview/photoview/api/test_utils"
)

func TestMain(m *testing.M) {
	os.Exit(test_utils.UnitTestRun(m))
}
