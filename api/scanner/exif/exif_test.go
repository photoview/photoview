// +build integration

package exif

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const bird_path = "./test_data/bird.jpg"

func TestExifParsers(t *testing.T) {
	t.Parallel()

	parsers := []struct {
		name   string
		parser exifParser
	}{
		{
			name:   "internal",
			parser: &internalExifParser{},
		},
		{
			name:   "external",
			parser: &externalExifParser{},
		},
	}

	for _, p := range parsers {
		t.Run(p.name, func(t *testing.T) {
			t.Parallel()

			exif, err := p.parser.ParseExif(bird_path)

			if assert.NoError(t, err) {
				assert.EqualValues(t, *exif.Camera, "Canon EOS 600D")
				assert.EqualValues(t, *exif.Maker, "Canon")
				assert.WithinDuration(t, *exif.DateShot, time.Unix(1336318784, 0).UTC(), time.Minute)
				assert.EqualValues(t, *exif.Aperture, 6.3)
				assert.EqualValues(t, *exif.Iso, 800)
				assert.EqualValues(t, *exif.FocalLength, 300)
				assert.EqualValues(t, *exif.Flash, 16)
				assert.EqualValues(t, *exif.Orientation, 1)
				assert.InDelta(t, *exif.GPSLatitude, 65.01681388888889, 0.0001)
				assert.InDelta(t, *exif.GPSLongitude, 25.466863888888888, 0.0001)
			}
		})
	}
}

// func TestExternalExifParser(t *testing.T) {
// 	parser := externalExifParser{}

// 	exif, err := parser.ParseExif((bird_path))

// 	if assert.NoError(t, err) {
// 		assert.Equal(t, exif, &bird_exif)
// 	}
// }
