package exif

import (
	"testing"

	_ "github.com/photoview/photoview/api/test_utils/flags"
)

func TestParseWithoutInit(t *testing.T) {
	if _, err := Parse("./test_data/bird.jpg"); err == nil {
		t.Fatalf("Parse() without Init() doesn't return an error")
	}
}

func TestParse(t *testing.T) {
	cleanup, err := Initialize(t.Context())
	if err != nil {
		t.Fatalf("Initialize() error: %v", err)
	}
	defer cleanup(t.Context())

	if _, err := Parse("./test_data/bird.jpg"); err != nil {
		t.Fatalf("Parse() returns an error: %v", err)
	}
}
