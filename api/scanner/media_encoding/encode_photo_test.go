package media_encoding

import (
	"testing"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner/media_type"
	_ "github.com/photoview/photoview/api/test_utils/flags"
)

// A caller which already knows the media type (e.g. from the album scanner
// cache) must be able to hand it over, otherwise every media triggers a second
// media type lookup, which shells out to exiftool.
func TestNewEncodeMediaDataUsesProvidedContentType(t *testing.T) {
	media := &models.Media{Path: "/does/not/exist.jpg"}

	data := NewEncodeMediaData(media, media_type.TypeJPEG)

	contentType, err := data.ContentType()
	if err != nil {
		t.Fatalf("ContentType() returned an unexpected error: %s", err)
	}

	if contentType != media_type.TypeJPEG {
		t.Errorf("ContentType() = %q, expected the provided %q", contentType.String(), media_type.TypeJPEG.String())
	}
}

// Passing TypeUnknown keeps the previous behaviour: the type is resolved lazily
// on the first ContentType() call.
func TestNewEncodeMediaDataResolvesUnknownContentTypeLazily(t *testing.T) {
	media := &models.Media{Path: "/does/not/exist.jpg"}

	data := NewEncodeMediaData(media, media_type.TypeUnknown)

	if data._contentType != media_type.TypeUnknown {
		t.Errorf("expected the content type to stay unresolved, got %q", data._contentType.String())
	}

	// The path does not exist, so resolving it cannot succeed.
	if _, err := data.ContentType(); err == nil {
		t.Error("expected ContentType() to report an error for an unresolvable media type")
	}
}
