package orient

import (
	"fmt"
	"strings"
	"testing"

	"github.com/photoview/photoview/api/scanner/externaltools/exiftool"
	"github.com/photoview/photoview/api/test_utils"
)

func TestMain(m *testing.M) {
	test_utils.UnitTestRun(m)
}

func TestEnsureExifOrient(t *testing.T) {
	et, err := exiftool.New()
	if err != nil {
		t.Fatalf("create exiftool error: %v", err)
	}
	defer et.Close()

	t.Log("Orientation explaination:")
	t.Log("1 = Horizontal (normal)")
	t.Log("2 = Mirror horizontal")
	t.Log("3 = Rotate 180")
	t.Log("4 = Mirror vertical")
	t.Log("5 = Mirror horizontal and rotate 270 CW")
	t.Log("6 = Rotate 90 CW")
	t.Log("7 = Mirror horizontal and rotate 90 CW")
	t.Log("8 = Rotate 270 CW")

	// Test files should be present in the same directory as this test
	for _, file := range []string{
		"left_arrow_normal_web.jpg",
		"up_arrow_90cw_web.jpg",
		"left_arrow_normal_nonweb.tiff",
		"up_arrow_90cw_nonweb.tiff",
	} {
		var meta exiftool.PhotoMeta
		if err := et.QueryJSONTagsByNumber(file, &meta); err != nil {
			t.Fatalf("exiftool.QueryJSONTagsByNumber() error: %v", err)
		}

		got := meta.Orientation
		want := int64(1)
		if strings.Contains(file, "_90cw_") {
			want = 6
		}

		if got == nil || *got != want {
			t.Errorf("file %q orientation is %v, want: %d", file, pointerString(got), want)
		}
	}
}

func pointerString[T any](v *T) string {
	if v == nil {
		return "(nil)"
	}

	return fmt.Sprintf("&(%v)", *v)
}
