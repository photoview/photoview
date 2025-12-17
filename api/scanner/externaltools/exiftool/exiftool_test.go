package exiftool

import (
	"fmt"
	"maps"
	"path/filepath"
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"

	_ "github.com/photoview/photoview/api/test_utils/flags"
)

func TestInstance(t *testing.T) {
	instance, err := New()
	if err != nil {
		t.Fatalf("create instance error: %v", err)
	}

	t.Log("bin:", instance.Binary())
	t.Log("version:", instance.Version())

	if instance.Binary() == "" {
		t.Errorf("want exiftool binary, but got an emtpy string")
	}

	if instance.Version() == "" {
		t.Errorf("want exiftool version, but got an emtpy string")
	}

	if err := instance.Close(); err != nil {
		t.Errorf("close instance error: %v", err)
	}
}

func TestQueryMIMEType(t *testing.T) {
	tests := []struct {
		file string
		want string
	}{
		{"./test_data/bird.jpg", "image/jpeg"},
		{"./test_data/exif_subsec_timezone.heic", "image/heic"},
		{"./test_data/cr3.cr3", "image/x-canon-cr3"},
	}

	instance, err := New()
	if err != nil {
		t.Fatalf("new error: %v", err)
	}
	defer instance.Close()

	for _, tc := range tests {
		t.Run(tc.file, func(t *testing.T) {
			got, err := instance.QueryMIMEType(tc.file)
			if err != nil {
				t.Errorf("QueryMIMEType(%q) error: %v", tc.file, err)
				return
			}

			if got != tc.want {
				t.Errorf("QueryMIMEType(%q) = %q, want: %q", tc.file, got, tc.want)
			}
		})
	}
}

func TestQueryTime(t *testing.T) {
	tests := []struct {
		file     string
		wantKeys []string
	}{
		{"./test_data/bird.jpg", []string{
			"DateCreated",
			"DateTimeCreated",
			"DateTimeOriginal",
			"FileAccessDate",
			"FileInodeChangeDate",
			"FileModifyDate",
			"TimeCreated",
		}},
		{"./test_data/cr3.cr3", []string{
			"CreateDate",
			"DateTimeOriginal",
			"DaylightSavings",
			"FileAccessDate",
			"FileInodeChangeDate",
			"FileModifyDate",
			"MediaCreateDate",
			"MediaModifyDate",
			"ModifyDate",
			"OffsetTime",
			"OffsetTimeDigitized",
			"OffsetTimeOriginal",
			"SubSecCreateDate",
			"SubSecDateTimeOriginal",
			"SubSecModifyDate",
			"TimeStamp",
			"TimeZone",
			"TimeZoneCity",
			"TrackCreateDate",
			"TrackModifyDate",
		}},
		{"./test_data/stripped.jpg", []string{
			"FileAccessDate",
			"FileInodeChangeDate",
			"FileModifyDate",
		}},
	}

	instance, err := New()
	if err != nil {
		t.Fatalf("new error: %v", err)
	}
	defer instance.Close()

	for _, tc := range tests {
		t.Run(tc.file, func(t *testing.T) {
			got, err := instance.QueryTime(tc.file)
			if err != nil {
				t.Errorf("QueryTime(%q) error: %v", tc.file, err)
				return
			}

			gotKeys := slices.Collect(maps.Keys(got))
			slices.Sort(gotKeys)
			if diff := cmp.Diff(gotKeys, tc.wantKeys); diff != "" {
				t.Errorf("QueryTime(%q), keys diff: (-got, +want)\n%s", tc.file, diff)
			}
		})
	}
}

func TestQueryGPS(t *testing.T) {
	tests := []struct {
		file              string
		wantLat, wantLong float64
	}{
		{"./test_data/CorrectGPS.jpg", 44.4789972, 11.2979222},
	}

	instance, err := New()
	if err != nil {
		t.Fatalf("new error: %v", err)
	}
	defer instance.Close()

	for _, tc := range tests {
		t.Run(tc.file, func(t *testing.T) {
			gotLat, gotLong, err := instance.QueryGPS(tc.file)
			if err != nil {
				t.Errorf("QueryGPS(%q) error: %v", tc.file, err)
				return
			}

			gpsToString := func(latitude, longitude float64) string {
				return fmt.Sprintf("(%.7f, %.7f)", latitude, longitude)
			}

			if got, want := gpsToString(gotLat, gotLong), gpsToString(tc.wantLat, tc.wantLong); got != want {
				t.Errorf("QUeryGPS(%q) = %s, want: %s", tc.file, got, want)
			}
		})
	}
}

func TestSaveJPEGPreview(t *testing.T) {
	tests := []struct {
		file   string
		output string
	}{
		{"./test_data/cr3.cr3", "cr3.jpg"},
	}

	instance, err := New()
	if err != nil {
		t.Fatalf("new error: %v", err)
	}
	defer instance.Close()

	outputDir := t.TempDir()

	for _, tc := range tests {
		t.Run(tc.file, func(t *testing.T) {
			output := filepath.Join(outputDir, tc.output)
			err := instance.SaveJPEGPreview(tc.file, output)
			if err != nil {
				t.Errorf("SaveJPEGPreview(%q, %q) error: %v", tc.file, output, err)
				return
			}

			mime, err := instance.QueryMIMEType(output)
			if err != nil {
				t.Errorf("QueryMIMEType(%q) error: %v", output, err)
				return
			}

			if got, want := mime, "image/jpeg"; got != want {
				t.Errorf("QueryMIMEType(%q) = %q, want: %q", output, got, want)
				return
			}
		})
	}

}
