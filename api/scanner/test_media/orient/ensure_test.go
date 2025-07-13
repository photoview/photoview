package orient

import (
	"os"
	"strings"
	"testing"

	"github.com/barasher/go-exiftool"
	"github.com/photoview/photoview/api/test_utils"
)

func TestMain(m *testing.M) {
	os.Exit(test_utils.IntegrationTestRun(m))
}

func TestEnsureExifOrient(t *testing.T) {
	buf := make([]byte, 64*1024)

	et, err := exiftool.NewExiftool(exiftool.NoPrintConversion(), exiftool.Buffer(buf, 64*1024))
	if err != nil {
		t.Fatalf("create exiftool error: %v", err)
	}
	defer et.Close()

	// Test files should be present in the same directory as this test
	for _, file := range []string{
		"left_arrow_normal_web.jpg",
		"up_arrow_90cw_web.jpg",
		"left_arrow_normal_nonweb.tiff",
		"up_arrow_90cw_nonweb.tiff",
	} {
		meta := et.ExtractMetadata(file)
		if got, want := len(meta), 1; got != want {
			t.Fatalf("len(file(%s) meta) = %d, want: %d", file, got, want)
		}

		got, err := meta[0].GetInt("Orientation")
		if err != nil {
			t.Fatalf("get orientation with file %s error: %v", file, err)
		}

		want := int64(1) // Normal
		if strings.Contains(file, "_90cw_") {
			want = 6 // 90 clockwise
		}

		if got != want {
			t.Errorf("file %q orientation is %d, want: %d", file, got, want)
		}
	}
}
