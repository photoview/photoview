package exif

import (
	"sync"
	"testing"

	_ "github.com/photoview/photoview/api/test_utils/flags"
	"github.com/spf13/afero"
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
	fs := afero.NewOsFs()

	if _, err := Parse(fs, "./test_data/bird.jpg"); err == nil {
		t.Fatalf("Parse() without Init() doesn't return an error")
	}
}

func TestParse(t *testing.T) {
	resetForTest()
	fs := afero.NewOsFs()

	cleanup, err := Initialize()
	if err != nil {
		t.Fatalf("Initialize() error: %v", err)
	}
	defer cleanup()

	filename := "./test_data/bird.jpg"

	metadata, err := Parse(fs, filename)
	if err != nil {
		t.Fatalf("Parse() returns an error: %v", err)
	}

	if metadata == nil {
		t.Errorf("Parse(%q) should not return nil", filename)
	}
}

func TestParseWithMemMapFs(t *testing.T) {
	resetForTest()
	osFs := afero.NewOsFs()
	memFs := afero.NewMemMapFs()

	cleanup, err := Initialize()
	if err != nil {
		t.Fatalf("Initialize() error: %v", err)
	}
	defer cleanup()

	// Copy test file from OS filesystem to memory filesystem
	filename := "./test_data/bird.jpg"
	data, err := afero.ReadFile(osFs, filename)
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}

	if err := memFs.MkdirAll("./test_data", 0o755); err != nil {
		t.Fatalf("Failed to create test_data dir in memFs: %v", err)
	}

	err = afero.WriteFile(memFs, filename, data, 0644)
	if err != nil {
		t.Fatalf("Failed to write test file to memFs: %v", err)
	}

	// Now test parsing from the memory filesystem
	metadata, err := Parse(memFs, filename)
	if err != nil {
		t.Fatalf("Parse() from MemMapFs returns an error: %v", err)
	}

	if metadata == nil {
		t.Errorf("Parse(%q) from MemMapFs should not return nil", filename)
	}
}

func TestMIMEType(t *testing.T) {
	resetForTest()
	fs := afero.NewOsFs()

	cleanup, err := Initialize()
	if err != nil {
		t.Fatalf("Initialize() error: %v", err)
	}
	defer cleanup()

	filename := "./test_data/bird.jpg"

	mime, err := MIMEType(fs, filename)
	if err != nil {
		t.Fatalf("MIMEType() returns an error: %v", err)
	}

	if mime == "" {
		t.Errorf("MIMEType(%q) should not return an empty string", filename)
	}
}
