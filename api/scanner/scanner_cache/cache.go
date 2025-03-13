package scanner_cache

import (
	"fmt"
	"log"
	"os"
	"path"
	"sync"

	"github.com/photoview/photoview/api/scanner/media_type"
	"github.com/photoview/photoview/api/scanner/scanner_utils"
)

type AlbumScannerCache struct {
	path_contains_photos map[string]bool
	photo_types          map[string]media_type.MediaType
	ignore_data          map[string][]string
	mutex                sync.Mutex
}

func MakeAlbumCache() *AlbumScannerCache {
	return &AlbumScannerCache{
		path_contains_photos: make(map[string]bool),
		photo_types:          make(map[string]media_type.MediaType),
		ignore_data:          make(map[string][]string),
	}
}

// Insert single album directory in cache
func (c *AlbumScannerCache) InsertAlbumPath(path string, containsPhoto bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.path_contains_photos[path] = containsPhoto
}

// Insert album path and all parent directories up to the given root directory in cache
func (c *AlbumScannerCache) InsertAlbumPaths(endPath string, root string, containsPhoto bool) {
	currPath := path.Clean(endPath)
	rootPath := path.Clean(root)

	c.mutex.Lock()
	defer c.mutex.Unlock()

	for currPath != rootPath || currPath == "." {

		c.path_contains_photos[currPath] = containsPhoto

		currPath = path.Dir(currPath)
	}
}

func (c *AlbumScannerCache) AlbumContainsPhotos(path string) *bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	contains_photo, found := c.path_contains_photos[path]
	if found {
		// log.Printf("Album cache hit: %s\n", path)
		return &contains_photo
	}

	return nil
}

// func (c *AlbumScannerCache) InsertPhotoType(path string, content_type MediaType) {
// 	c.mutex.Lock()
// 	defer c.mutex.Unlock()

// 	(c.photo_types)[path] = content_type
// }

func (c *AlbumScannerCache) GetMediaType(path string) (media_type.MediaType, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	result, found := c.photo_types[path]
	if found {
		return result, nil
	}

	mediaType := media_type.GetMediaType(path)
	if mediaType == media_type.TypeUnknown {
		return mediaType, fmt.Errorf("unknown media type (%s)", path)
	}

	c.photo_types[path] = mediaType

	return mediaType, nil
}

func (c *AlbumScannerCache) GetAlbumIgnore(path string) *[]string {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	ignore_data, found := c.ignore_data[path]
	if found {
		return &ignore_data
	}

	return nil
}

func (c *AlbumScannerCache) InsertAlbumIgnore(path string, ignoreData []string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.ignore_data[path] = ignoreData
}

func (c *AlbumScannerCache) IsPathMedia(mediaPath string) bool {
	mediaType, err := c.GetMediaType(mediaPath)
	if err != nil {
		scanner_utils.ScannerError("IsPathMedia (%s): %s", mediaPath, err)
		return false
	}

	// Ignore hidden files
	if path.Base(mediaPath)[0:1] == "." {
		return false
	}

	if !mediaType.IsSupported() {
		log.Printf("Unsupported media type %q for file: %s\n", mediaType, mediaPath)
		return false
	}

	// Make sure file isn't empty
	fileStats, err := os.Stat(mediaPath)
	if err != nil || fileStats.Size() == 0 {
		return false
	}

	return true
}
