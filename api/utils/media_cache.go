package utils

import (
	"os"
	"path"
	"strconv"

	"github.com/pkg/errors"
)

// CachePathForMedia is a low-level implementation for Media.CachePath()
func CachePathForMedia(albumID int, mediaID int) (string, error) {

	// Make root cache dir if not exists
	if _, err := os.Stat(MediaCachePath()); os.IsNotExist(err) {
		if err := os.Mkdir(MediaCachePath(), os.ModePerm); err != nil {
			return "", errors.Wrap(err, "could not make root image cache directory")
		}
	}

	// Make album cache dir if not exists
	albumCachePath := path.Join(MediaCachePath(), strconv.Itoa(int(albumID)))
	if _, err := os.Stat(albumCachePath); os.IsNotExist(err) {
		if err := os.Mkdir(albumCachePath, os.ModePerm); err != nil {
			return "", errors.Wrap(err, "could not make album image cache directory")
		}
	}

	// Make photo cache dir if not exists
	photoCachePath := path.Join(albumCachePath, strconv.Itoa(int(mediaID)))
	if _, err := os.Stat(photoCachePath); os.IsNotExist(err) {
		if err := os.Mkdir(photoCachePath, os.ModePerm); err != nil {
			return "", errors.Wrap(err, "could not make photo image cache directory")
		}
	}

	return photoCachePath, nil
}

var test_cache_path string = ""

func ConfigureTestCache(tmp_dir string) {
	test_cache_path = tmp_dir
}

// MediaCachePath returns the path for where the media cache is located on the file system
func MediaCachePath() string {
	if test_cache_path != "" {
		return test_cache_path
	}

	photoCache := EnvMediaCachePath.GetValue()
	if photoCache == "" {
		photoCache = "./media_cache"
	}

	return photoCache
}
