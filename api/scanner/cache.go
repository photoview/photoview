package scanner

import "path"

type AlbumScannerCache struct {
	path_contains_photos map[string]bool
	photo_types          map[string]ImageType
}

func MakeAlbumCache() *AlbumScannerCache {
	return &AlbumScannerCache{
		path_contains_photos: make(map[string]bool),
		photo_types:          make(map[string]ImageType),
	}
}

// Insert single album directory in cache
func (c *AlbumScannerCache) InsertAlbumPath(path string, contains_photo bool) {
	c.path_contains_photos[path] = contains_photo
}

// Insert album path and all parent directories up to the given root directory in cache
func (c *AlbumScannerCache) InsertAlbumPaths(end_path string, root string, contains_photo bool) {
	curr_path := path.Clean(end_path)
	root_path := path.Clean(root)

	for curr_path != root_path || curr_path == "." {

		c.InsertAlbumPath(curr_path, contains_photo)

		curr_path = path.Dir(curr_path)
	}
}

func (c *AlbumScannerCache) AlbumContainsPhotos(path string) *bool {
	contains_photo, found := c.path_contains_photos[path]
	if found {
		// log.Printf("Album cache hit: %s\n", path)
		return &contains_photo
	}

	return nil
}

func (c *AlbumScannerCache) InsertPhotoType(path string, content_type ImageType) {
	(c.photo_types)[path] = content_type
}

func (c *AlbumScannerCache) GetPhotoType(path string) *ImageType {
	result, found := c.photo_types[path]
	if found {
		// log.Printf("Image cache hit: %s\n", path)
		return &result
	}

	return nil
}
