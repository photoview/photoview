package exif_test

import (
	"fmt"
	"os"
	"path"
	"testing"
	"time"

	"github.com/barasher/go-exiftool"
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner/exif"
	"github.com/photoview/photoview/api/test_utils"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	os.Exit(test_utils.IntegrationTestRun(m))
}

func TestExifParsers(t *testing.T) {
	test_utils.FilesystemTest(t)

	parsers := []struct {
		name   string
		parser exif.ExifParser
	}{
		{
			name:   "internal",
			parser: exif.NewInternalExifParser(),
		},
	}

	if externalParser, err := exif.NewExiftoolParser(); err == nil {
		parsers = append(parsers, struct {
			name   string
			parser exif.ExifParser
		}{
			name:   "external",
			parser: externalParser,
		})
	}

	images := []struct {
		path   string
		assert func(t *testing.T, exif *models.MediaEXIF)
	}{
		{
			path: "./test_data/bird.jpg",
			assert: func(t *testing.T, exif *models.MediaEXIF) {
				assert.EqualValues(t, *exif.Description, "Photo of a Bird")
				assert.WithinDuration(t, *exif.DateShot, time.Unix(1336318784, 0).UTC(), time.Minute)
				assert.EqualValues(t, *exif.Camera, "Canon EOS 600D")
				assert.EqualValues(t, *exif.Maker, "Canon")
				assert.WithinDuration(t, *exif.DateShot, time.Unix(1336318784, 0).UTC(), time.Minute)
				assert.InDelta(t, *exif.Exposure, 1.0/4000.0, 0.0001)
				assert.EqualValues(t, *exif.Aperture, 6.3)
				assert.EqualValues(t, *exif.Iso, 800)
				assert.EqualValues(t, *exif.FocalLength, 300)
				assert.EqualValues(t, *exif.Flash, 16)
				assert.EqualValues(t, *exif.Orientation, 1)
				assert.InDelta(t, *exif.GPSLatitude, 65.01681388888889, 0.0001)
				assert.InDelta(t, *exif.GPSLongitude, 25.466863888888888, 0.0001)
			},
		},
		{
			path: "./test_data/stripped.jpg",
			assert: func(t *testing.T, exif *models.MediaEXIF) {
				assert.Nil(t, exif)
			},
		},
		{
			path: "./test_data/bad-exif.jpg",
			assert: func(t *testing.T, exif *models.MediaEXIF) {
				assert.Nil(t, exif.Exposure)
			},
		},
	}

	for _, p := range parsers {
		for _, img := range images {
			t.Run(fmt.Sprintf("%s:%s", p.name, path.Base(img.path)), func(t *testing.T) {

				if p.name == "external" {
					_, err := exiftool.NewExiftool()
					if err != nil {
						t.Skip("failed to get exiftool, skipping test")
					}
				}

				exif, err := p.parser.ParseExif(img.path)

				if assert.NoError(t, err) {
					img.assert(t, exif)
				}
			})
		}
	}
}

// func TestExternalExifParser(t *testing.T) {
// 	parser := externalExifParser{}

// 	exif, err := parser.ParseExif((bird_path))

// 	if assert.NoError(t, err) {
// 		assert.Equal(t, exif, &bird_exif)
// 	}
// }
