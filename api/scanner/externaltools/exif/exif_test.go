package exif

import (
	"os"
	"sync"
	"testing"
	"time"

	_ "github.com/photoview/photoview/api/test_utils/flags"
)

func resetForTest() {
	globalMu.Lock()
	defer globalMu.Unlock()
	if globalExifParser != nil {
		_ = globalExifParser.Close()
	}
	globalExifParser = nil
	// Allow Initialize() to run again in subsequent tests
	globalInit = sync.Once{}
}

func TestParseWithoutInit(t *testing.T) {
	resetForTest()
	if _, err := Parse("./test_data/bird.jpg"); err == nil {
		t.Fatalf("Parse() without Init() doesn't return an error")
	}
}

func TestParse(t *testing.T) {
	resetForTest()

	cleanup, err := Initialize()
	if err != nil {
		t.Fatalf("Initialize() error: %v", err)
	}
	defer cleanup()

	filename := "./test_data/bird.jpg"

	metadata, err := Parse(filename)
	if err != nil {
		t.Fatalf("Parse() returns an error: %v", err)
	}

	if metadata == nil {
		t.Errorf("Parse(%q) should not return nil", filename)
	}
}

func TestMIMEType(t *testing.T) {
	resetForTest()

	cleanup, err := Initialize()
	if err != nil {
		t.Fatalf("Initialize() error: %v", err)
	}
	defer cleanup()

	filename := "./test_data/bird.jpg"

	mime, err := MIMEType(filename)
	if err != nil {
		t.Fatalf("MIMEType() returns an error: %v", err)
	}

	if mime == "" {
		t.Errorf("MIMEType(%q) should not return an empty string", filename)
	}
}

func fileModifyDateLiteralInUTC(t *testing.T, file string) time.Time {
	fstat, err := os.Stat(file)
	if err != nil {
		t.Fatalf("os.Stat(%q) error: %v", file, err)
	}

	ret := fstat.ModTime().Truncate(time.Second)
	_, offset := ret.Zone()
	ret = ret.Add(time.Duration(offset) * time.Second).UTC()

	return ret
}

func mustParseNoTimeZone(t *testing.T, timeStr string) time.Time {
	layout := "2006:01:02 15:04:05.999"
	ret, err := time.ParseInLocation(layout, timeStr, time.UTC)
	if err != nil {
		t.Fatalf("time.Parse(%q) error: %v", timeStr, err)
	}
	return ret
}

func TestSamplesTime(t *testing.T) {
	resetForTest()

	cleanup, err := Initialize()
	if err != nil {
		t.Fatalf("Initialize() error: %v", err)
	}
	defer cleanup()

	tests := []struct {
		file          string
		wantTime      time.Time
		wantOffsetSec *int
	}{
		{"./test_data/sample1.heif", fileModifyDateLiteralInUTC(t, "./test_data/sample1.heif"), nil},
		{"./test_data/sample1_nef.jpg", mustParseNoTimeZone(t, "2008:03:15 07:44:21.49"), new(-7 * 60 * 60)},
	}

	for _, tc := range tests {
		t.Run(tc.file, func(t *testing.T) {
			metadata, err := Parse(tc.file)
			if err != nil {
				t.Fatalf("Parse(%q) returns an error: %v", tc.file, err)
			}

			if metadata == nil {
				t.Fatalf("metadata == nil")
			}
			if metadata.DateShot == nil {
				t.Fatalf("metadata.DateShot == nil")
			}

			if got := *metadata.DateShot; !got.Equal(tc.wantTime) {
				t.Errorf("metadata.DateShot = %v, want: %v", got, tc.wantTime)
			}

			if tc.wantOffsetSec == nil {
				if got := metadata.OffsetSecShot; got != nil {
					t.Errorf("metadata.OffsetSecShot = %+v, want: nil", got)
				}
			} else {
				if got, want := *metadata.OffsetSecShot, *tc.wantOffsetSec; got != want {
					t.Errorf("metadata.OffsetSecShot = &(%v), want: &(%v)", got, want)
				}
			}
		})
	}
}
