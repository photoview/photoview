package queue

import (
	"os"
	"path"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner/scanner_cache"
	"github.com/photoview/photoview/api/utils"
	"gorm.io/gorm"
)

// AlbumRequest asks the queue to (re)scan a single album directory.
type AlbumRequest struct {
	Album *models.Album
	Cache *scanner_cache.AlbumScannerCache
}

// expandAlbum lists the direct contents of an album directory and builds one
// task per candidate media file. It deliberately does no database or
// exif/ffmpeg work itself - that happens per-task, in gather(), spread across
// workers instead of serializing it all up front in the dispatcher.
func expandAlbum(db *gorm.DB, req AlbumRequest) ([]*task, *albumState, error) {
	album := req.Album
	cache := req.Cache

	dirContent, err := os.ReadDir(album.Path)
	if err != nil {
		return nil, nil, err
	}

	var paths []string
	for _, item := range dirContent {
		mediaPath := path.Join(album.Path, item.Name())

		isDirSymlink, err := utils.IsDirSymlink(mediaPath)
		if err != nil {
			isDirSymlink = false
		}

		if item.IsDir() || isDirSymlink {
			continue
		}

		if !cache.IsPathMedia(mediaPath) {
			continue
		}

		paths = append(paths, mediaPath)
	}

	state := newAlbumState(db, album, cache, len(paths))

	tasks := make([]*task, len(paths))
	for i, p := range paths {
		tasks[i] = newTask(album, cache, state, p, nil)
	}

	return tasks, state, nil
}
