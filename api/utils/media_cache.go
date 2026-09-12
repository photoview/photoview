package utils

import (
	"os"
	"path"
	"strconv"
	"sync"

	"github.com/pkg/errors"
)

// mkdirIfNotExists creates dir, tolerating a concurrent caller having
// already created it (os.Stat-then-os.Mkdir would be a TOCTOU race under
// concurrent workers sharing the same album - this checks the actual Mkdir
// outcome instead of pre-checking existence).
func mkdirIfNotExists(dir string) error {
	if err := os.Mkdir(dir, os.ModePerm); err != nil && !os.IsExist(err) {
		return err
	}
	return nil
}

// CachePathForMedia is a low-level implementation for Media.CachePath()
func CachePathForMedia(albumID int, mediaID int) (string, error) {

	// Make root cache dir if not exists
	if err := mkdirIfNotExists(MediaCachePath()); err != nil {
		return "", errors.Wrap(err, "could not make root image cache directory")
	}

	// Make album cache dir if not exists
	albumCachePath := path.Join(MediaCachePath(), strconv.Itoa(int(albumID)))
	if err := mkdirIfNotExists(albumCachePath); err != nil {
		return "", errors.Wrap(err, "could not make album image cache directory")
	}

	// Make photo cache dir if not exists
	photoCachePath := path.Join(albumCachePath, strconv.Itoa(int(mediaID)))
	if err := mkdirIfNotExists(photoCachePath); err != nil {
		return "", errors.Wrap(err, "could not make photo image cache directory")
	}

	return photoCachePath, nil
}

// PendingCachePathForMedia returns (and creates) a scratch cache directory
// for a media file that hasn't been assigned a database ID yet - the
// scanner writes generated files here before a row exists, then renames
// this into the real CachePathForMedia location once persisted.
func PendingCachePathForMedia(albumID int, key string) (string, error) {
	// Make root cache dir if not exists
	if err := mkdirIfNotExists(MediaCachePath()); err != nil {
		return "", errors.Wrap(err, "could not make root image cache directory")
	}

	// Make album cache dir if not exists
	albumCachePath := path.Join(MediaCachePath(), strconv.Itoa(albumID))
	if err := mkdirIfNotExists(albumCachePath); err != nil {
		return "", errors.Wrap(err, "could not make album image cache directory")
	}

	pendingPath := path.Join(albumCachePath, "pending-"+key)
	if err := os.MkdirAll(pendingPath, os.ModePerm); err != nil {
		return "", errors.Wrap(err, "could not make pending media cache directory")
	}

	return pendingPath, nil
}

// MediaCacheLeafPath returns the final cache directory path for a media -
// same location as CachePathForMedia, but as a pure path computation: it
// does not create anything (not even the album directory). This is what
// PendingCachePathForMedia's scratch directory gets renamed onto, since
// os.Rename needs the destination to not already exist.
func MediaCacheLeafPath(albumID int, mediaID int) string {
	return path.Join(MediaCachePath(), strconv.Itoa(albumID), strconv.Itoa(mediaID))
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

var cachedMediaCachePath = sync.OnceValue(func() string {
	return computeMediaCachePath()
})

// MediaCachePath returns the path for where the media cache is located on the file system
func MediaCachePath() string {
	testCachePathLocker.RLock()
	cachedPath := testCachePath
	testCachePathLocker.RUnlock()
	if cachedPath != "" {
		return cachedPath
	}

	return cachedMediaCachePath()
}

func computeMediaCachePath() string {
	photoCache := EnvMediaCachePath.GetValue()
	if photoCache == "" {
		photoCache = "./media_cache"
	}
	return photoCache
}
