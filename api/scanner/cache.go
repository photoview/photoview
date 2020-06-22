package scanner

import "path"

type ScannerCache struct {
	cache               map[string]interface{}
	photo_paths_scanned []interface{}
	album_paths_scanned []interface{}
}

func MakeScannerCache() ScannerCache {
	return ScannerCache{
		cache:               make(map[string]interface{}),
		photo_paths_scanned: make([]interface{}, 0),
		album_paths_scanned: make([]interface{}, 0),
	}
}

func (c *ScannerCache) insert_photo_type(path string, content_type ImageType) {
	(c.cache)["photo_type//"+path] = content_type
}

func (c *ScannerCache) get_photo_type(path string) *string {
	result, found := (c.cache)["photo_type//"+path].(string)
	if found {
		// log.Printf("Image cache hit: %s\n", path)
		return &result
	}

	return nil
}

// Insert single album directory in cache
func (c *ScannerCache) insert_album_path(path string, contains_photo bool) {
	(c.cache)["album_path//"+path] = contains_photo
}

// Insert album path and all parent directories up to the given root directory in cache
func (c *ScannerCache) insert_album_paths(end_path string, root string, contains_photo bool) {
	curr_path := path.Clean(end_path)
	root_path := path.Clean(root)

	for curr_path != root_path || curr_path == "." {

		c.insert_album_path(curr_path, contains_photo)

		curr_path = path.Dir(curr_path)
	}
}

func (c *ScannerCache) album_contains_photo(path string) *bool {
	contains_photo, found := (c.cache)["album_path//"+path].(bool)
	if found {
		// log.Printf("Album cache hit: %s\n", path)
		return &contains_photo
	}

	return nil
}
