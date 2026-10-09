package scanner_cache

import (
	"fmt"
	"sync"
	"testing"

	_ "github.com/photoview/photoview/api/test_utils/flags"
)

// The album scanner cache is shared by all scanner jobs, so every accessor has
// to stay safe for concurrent use. GetMediaType in particular resolves the media
// type (which shells out to exiftool) outside of the cache mutex now, so this
// test guards against reintroducing an unsynchronised map access.
//
// Run with `go test -race` to make it meaningful.
func TestAlbumScannerCacheConcurrentAccess(t *testing.T) {
	cache := MakeAlbumCache()

	const workers = 32

	var wg sync.WaitGroup
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func(i int) {
			defer wg.Done()

			// A path every worker touches, to force contention on the cache.
			shared := "/media/shared"
			cache.GetMediaType(shared + ".jpg")
			cache.InsertAlbumPath(shared, true)
			_ = cache.AlbumContainsPhotos(shared)
			cache.InsertAlbumIgnore(shared, []string{"*.tmp"})
			_ = cache.GetAlbumIgnore(shared)

			// A path exclusive to this worker, so the assertion is deterministic.
			unique := fmt.Sprintf("/media/unique-%d", i)
			cache.InsertAlbumPath(unique, true)

			containsPhotos := cache.AlbumContainsPhotos(unique)
			if containsPhotos == nil {
				t.Errorf("expected %q to be present in the album cache", unique)
				return
			}
			if !*containsPhotos {
				t.Errorf("expected %q to be marked as containing photos", unique)
			}
		}(i)
	}

	wg.Wait()
}

// Unknown media types are not cached, so they are looked up again on the next
// call. This is pre-existing behaviour and must not change with the narrower
// locking.
func TestAlbumScannerCacheUnknownMediaTypeIsNotCached(t *testing.T) {
	cache := MakeAlbumCache()

	if got := cache.GetMediaType("/does/not/exist.jpg"); got.String() != "unknown" {
		t.Errorf("expected an unknown media type, got %q", got.String())
	}

	if _, found := cache.photo_types["/does/not/exist.jpg"]; found {
		t.Error("expected unknown media types not to be stored in the cache")
	}
}
