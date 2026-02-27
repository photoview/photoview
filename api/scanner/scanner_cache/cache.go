package scanner_cache

import (
	"io/fs"
	"log"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/photoview/photoview/api/scanner/media_type"
	"github.com/photoview/photoview/api/utils"
)

type AlbumScannerCache struct {
	path_contains_photos map[string]bool
	photo_types          map[string]media_type.MediaType
	ignore_data          map[string][]string
	skip_extensions      map[string]struct{}
	mutex                sync.Mutex
}

func splitScannerSkipList(raw string) []string {
	return strings.FieldsFunc(raw, func(r rune) bool {
		switch r {
		case ',', ';', ' ', '\n', '\r', '\t':
			return true
		default:
			return false
		}
	})
}

func normalizeSkipExtension(ext string) string {
	normalized := strings.ToLower(strings.TrimSpace(ext))
	if normalized == "" {
		return ""
	}

	if !strings.HasPrefix(normalized, ".") {
		normalized = "." + normalized
	}

	return normalized
}

func parseSkipExtensions(raw string) map[string]struct{} {
	extensions := make(map[string]struct{})
	for _, item := range splitScannerSkipList(raw) {
		ext := normalizeSkipExtension(item)
		if ext == "" {
			continue
		}

		extensions[ext] = struct{}{}
	}

	return extensions
}

func (c *AlbumScannerCache) shouldSkipByConfiguredExtension(mediaPath string) bool {
	ext := strings.ToLower(path.Ext(mediaPath))
	if ext == "" {
		return false
	}

	_, ok := c.skip_extensions[ext]
	return ok
}

func (c *AlbumScannerCache) ShouldSkipMediaPath(mediaPath string) bool {
	if strings.HasPrefix(path.Base(mediaPath), ".") {
		return true
	}

	if c.shouldSkipByConfiguredExtension(mediaPath) {
		return true
	}

	return false
}

func MakeAlbumCache() *AlbumScannerCache {
	return &AlbumScannerCache{
		path_contains_photos: make(map[string]bool),
		photo_types:          make(map[string]media_type.MediaType),
		ignore_data:          make(map[string][]string),
		skip_extensions:      parseSkipExtensions(utils.EnvScannerSkipExtensions.GetValue()),
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

func (c *AlbumScannerCache) GetMediaType(path string) media_type.MediaType {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	result, found := c.photo_types[path]
	if found {
		return result
	}

	mediaType := media_type.GetMediaType(path)
	if mediaType == media_type.TypeUnknown {
		return mediaType
	}

	c.photo_types[path] = mediaType

	return mediaType
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

func (c *AlbumScannerCache) isPathMediaWithInfo(mediaPath string, fileInfo fs.FileInfo) bool {
	if fileInfo == nil || fileInfo.Size() == 0 {
		return false
	}

	mediaType := c.GetMediaType(mediaPath)
	if !mediaType.IsSupported() {
		log.Printf("Unsupported media type %q for file: %s\n", mediaType, mediaPath)
		return false
	}

	return true
}

func (c *AlbumScannerCache) IsPathMediaWithInfo(mediaPath string, fileInfo fs.FileInfo) bool {
	if c.ShouldSkipMediaPath(mediaPath) {
		return false
	}

	return c.isPathMediaWithInfo(mediaPath, fileInfo)
}

func (c *AlbumScannerCache) IsPathMedia(mediaPath string) bool {
	if c.ShouldSkipMediaPath(mediaPath) {
		return false
	}

	fileStats, err := os.Stat(mediaPath)
	if err != nil {
		return false
	}

	return c.isPathMediaWithInfo(mediaPath, fileStats)
}
