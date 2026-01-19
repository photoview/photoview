package utils

import (
	"os"
	"path"
	"strconv"
	"sync"

	"github.com/pkg/errors"
	"github.com/spf13/afero"
)

// CachePathForMedia is a low-level implementation for Media.CachePath()
func CachePathForMedia(cacheFs afero.Fs, albumID int, mediaID int) (string, error) {

	// Make root cache dir if not exists
	if _, err := cacheFs.Stat(MediaCachePath()); os.IsNotExist(err) {
		if err := cacheFs.Mkdir(MediaCachePath(), os.ModePerm); err != nil {
			return "", errors.Wrap(err, "could not make root image cache directory")
		}
	}

	// Make album cache dir if not exists
	albumCachePath := path.Join(MediaCachePath(), strconv.Itoa(int(albumID)))
	if _, err := cacheFs.Stat(albumCachePath); os.IsNotExist(err) {
		if err := cacheFs.Mkdir(albumCachePath, os.ModePerm); err != nil {
			return "", errors.Wrap(err, "could not make album image cache directory")
		}
	}

	// Make photo cache dir if not exists
	photoCachePath := path.Join(albumCachePath, strconv.Itoa(int(mediaID)))
	if _, err := cacheFs.Stat(photoCachePath); os.IsNotExist(err) {
		if err := cacheFs.Mkdir(photoCachePath, os.ModePerm); err != nil {
			return "", errors.Wrap(err, "could not make photo image cache directory")
		}
	}

	return photoCachePath, nil
}

var (
	testCachePath       string = ""
	testCachePathLocker sync.RWMutex
)

func GetTestCachePath() string {
	testCachePathLocker.RLock()
	defer testCachePathLocker.RUnlock()
	return testCachePath
}

func ConfigureTestCache(tmpDir string) {
	testCachePathLocker.Lock()
	defer testCachePathLocker.Unlock()
	testCachePath = tmpDir
}

// MediaCachePath returns the path for where the media cache is located on the file system
func MediaCachePath() string {
	testCachePathLocker.RLock()
	cachedPath := testCachePath
	testCachePathLocker.RUnlock()
	if cachedPath != "" {
		return cachedPath
	}

	photoCache := EnvMediaCachePath.GetValue()
	if photoCache == "" {
		photoCache = "./media_cache"
	}

	return photoCache
}
