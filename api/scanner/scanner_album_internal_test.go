package scanner

import (
	"context"
	"errors"
	"io/fs"
	"path"
	"testing"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner/scanner_cache"
	"github.com/photoview/photoview/api/scanner/scanner_task"
)

type testDirEntry struct {
	name    string
	info    fs.FileInfo
	infoErr error
}

func (e testDirEntry) Name() string {
	return e.name
}

func (e testDirEntry) IsDir() bool {
	return false
}

func (e testDirEntry) Type() fs.FileMode {
	return 0
}

func (e testDirEntry) Info() (fs.FileInfo, error) {
	return e.info, e.infoErr
}

func TestAlbumMediaInfoSkipsFilesWithInfoErrors(t *testing.T) {
	albumPath := t.TempDir()
	mediaPath := path.Join(albumPath, "broken.jpg")
	ctx := scanner_task.NewTaskContext(context.Background(), nil, &models.Album{Path: albumPath}, scanner_cache.MakeAlbumCache())

	itemInfo, isMedia := albumMediaInfo(ctx, testDirEntry{
		name:    "broken.jpg",
		infoErr: errors.New("permission denied"),
	}, mediaPath)

	if isMedia {
		t.Fatal("albumMediaInfo() should skip files when DirEntry.Info() fails")
	}

	if itemInfo != nil {
		t.Fatal("albumMediaInfo() should not return file info when DirEntry.Info() fails")
	}
}
