package resolvers

// Helper functions for upload.go, kept in a separate file (not the
// gqlgen-managed resolver file) since gqlgen's codegen doesn't recognize
// non-resolver top-level declarations and will otherwise comment them out
// on the next `go generate` as "unknown code".

import (
	"errors"
	"path/filepath"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner/scanner_queue"
)

// addAlbumToQueue is a seam over scanner_queue.AddAlbumToQueue so tests can
// avoid touching the real, process-global scanner queue (which starts a
// background worker goroutine on first use).
var addAlbumToQueue = scanner_queue.AddAlbumToQueue

func albumIDs(albums []*models.Album) []int {
	ids := make([]int, len(albums))
	for i, a := range albums {
		ids[i] = a.ID
	}
	return ids
}

func sanitizeFileName(name string) (string, error) {
	if name == "" {
		return "", errors.New("name must not be empty")
	}
	if name == "." || name == ".." {
		return "", errors.New("invalid name")
	}
	if filepath.Base(name) != name {
		return "", errors.New("name must not contain path separators")
	}
	return name, nil
}
