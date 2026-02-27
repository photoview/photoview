package scanner_cache

import (
	"testing"

	"github.com/photoview/photoview/api/utils"
)

func TestShouldSkipMediaPath(t *testing.T) {
	t.Setenv(utils.EnvScannerSkipExtensions.GetName(), ".gpx, KML ; geojson")

	cache := MakeAlbumCache()

	tests := []struct {
		path string
		want bool
	}{
		{path: "/photos/.hidden.jpg", want: true},
		{path: "/photos/track.gpx", want: true},
		{path: "/photos/track.KML", want: true},
		{path: "/photos/map.geojson", want: true},
		{path: "/photos/photo.jpg", want: false},
		{path: "/photos/raw.tiff", want: false},
	}

	for _, tc := range tests {
		if got := cache.ShouldSkipMediaPath(tc.path); got != tc.want {
			t.Errorf("ShouldSkipMediaPath(%q) = %v, want %v", tc.path, got, tc.want)
		}
	}
}
