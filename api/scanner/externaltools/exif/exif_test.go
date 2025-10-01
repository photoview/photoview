package exif

import (
	"sync"
	"testing"

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
