package exif

import (
	"fmt"
	"math"
	"path"
	"testing"
	"time"

	"github.com/barasher/go-exiftool"
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/stretchr/testify/assert"
)

func TestExifParsers(t *testing.T) {
	parser, err := NewExifParser()
	if err != nil {
		t.Fatalf("can't init exiftool: %v", err)
	}
	defer parser.Close()

	parsers := []struct {
		name   string
		parser *ExifParser
	}{
		{
			name:   "external",
			parser: parser,
		},
	}

	images := []struct {
		path   string
		assert func(t *testing.T, exif *models.MediaEXIF, err error)
	}{
		{
			path: "./test_data/bird.jpg",
			assert: func(t *testing.T, exif *models.MediaEXIF, err error) {
				assert.NoError(t, err)
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
			assert: func(t *testing.T, exif *models.MediaEXIF, err error) {
				assert.NoError(t, err)
				if exif == nil {
					assert.Nil(t, exif)
				} else {
					assert.Equal(t, 0, exif.ID)
					assert.True(t, exif.CreatedAt.IsZero())
					assert.True(t, exif.UpdatedAt.IsZero())
					assert.Nil(t, exif.Description)
					assert.Nil(t, exif.Camera)
					assert.Nil(t, exif.Maker)
					assert.Nil(t, exif.Lens)
					assert.Nil(t, exif.Exposure)
					assert.Nil(t, exif.Aperture)
					assert.Nil(t, exif.Iso)
					assert.Nil(t, exif.FocalLength)
					assert.Nil(t, exif.Flash)
					assert.Nil(t, exif.Orientation)
					assert.Nil(t, exif.ExposureProgram)
					assert.Nil(t, exif.GPSLatitude)
					assert.Nil(t, exif.GPSLongitude)
				}
			},
		},
		{
			path: "./test_data/bad-exif.jpg",
			assert: func(t *testing.T, exif *models.MediaEXIF, err error) {
				assert.NoError(t, err)
				assert.Nil(t, exif.Exposure)
			},
		},
		{
			path: "./test_data/IncorrectGPS.jpg",
			assert: func(t *testing.T, exif *models.MediaEXIF, err error) {
				assert.Nil(t, exif.GPSLatitude,
					"GPSLatitude expected to be NULL for an incorrect input data: %+v", exif.GPSLatitude)
				assert.Nil(t, exif.GPSLongitude,
					"GPSLongitude expected to be NULL for an incorrect input data: %+v", exif.GPSLongitude)
			},
		},
		{
			path: "./test_data/CorrectGPS.jpg",
			assert: func(t *testing.T, exif *models.MediaEXIF, err error) {
				const precision = 1e-7
				assert.NoError(t, err)
				assert.NotNil(t, exif.GPSLatitude,
					"GPSLatitude expected to be Not-NULL for a correct input data: %+v", exif.GPSLatitude)
				assert.NotNil(t, exif.GPSLongitude,
					"GPSLongitude expected to be Not-NULL for a correct input data: %+v", exif.GPSLongitude)
				assert.InDelta(t, *exif.GPSLatitude, 44.478997222222226, precision,
					"The exact value from input data is expected: %+v", exif.GPSLatitude)
				assert.InDelta(t, *exif.GPSLongitude, 11.297922222222223, precision,
					"The exact value from input data is expected: %+v", exif.GPSLongitude)
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

				img.assert(t, exif, err)
			})
		}
	}
}

func TestExtractValidGPSData(t *testing.T) {
	tests := []struct {
		name                string
		latitude, longitude float64
		wantOK              bool
	}{
		{"LatNormalLongNormal", 10.0, 10.0, true},

		{"LatNilLongNormal", math.NaN(), 10.0, false},
		{"LatNormalLongNil", 10.0, math.NaN(), false},

		{"Lat>90LongNormal", 100.0, 10.0, false},
		{"Lat<-90LongNormal", -100.0, 10.0, false},

		{"LatNormalLong>180", 10.0, 190.0, false},
		{"LatNormalLong<-180", 10.0, -190.0, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			metadata := exiftool.EmptyFileMetadata()
			if !math.IsNaN(tc.latitude) {
				metadata.SetFloat("GPSLatitude", tc.latitude)
			}
			if !math.IsNaN(tc.longitude) {
				metadata.SetFloat("GPSLongitude", tc.longitude)
			}

			lat, long, err := extractValidGPSData(&metadata)
			gotOK := err == nil
			if got, want := gotOK, tc.wantOK; got != want {
				t.Fatalf("extractValidGPSData({lat: %f, long: %f}) got an error: %v, want: %v", tc.latitude, tc.longitude, err, want)
			}

			if err != nil {
				// no need to check data if there is an error
				return
			}

			if got, want := lat, tc.latitude; got != want {
				t.Fatalf("extractValidGPSData({lat: %f, long: %f}) got latitude: %v, want: %v", tc.latitude, tc.longitude, got, want)
			}
			if got, want := long, tc.longitude; got != want {
				t.Fatalf("extractValidGPSData({lat: %f, long: %f}) got longitude: %v, want: %v", tc.latitude, tc.longitude, got, want)
			}
		})
	}
}

func TestIsFloatReal(t *testing.T) {
	tests := []struct {
		name  string
		value float64
		want  bool
	}{
		{"Normal", 10.0, true},
		{"+Inf", math.Inf(1), false},
		{"-Inf", math.Inf(-1), false},
		{"Nan", math.NaN(), false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := isFloatReal(tc.value)
			if got != tc.want {
				t.Errorf("isFloatReal(%f) = %v, want: %v", tc.value, got, tc.want)
			}
		})
	}
}

func TestSanitizeEXIF(t *testing.T) {
	nan := math.NaN()
	var exif models.MediaEXIF

	tests := []struct {
		field string
		ptr   **float64
	}{
		{"Exposure", &exif.Exposure},
		{"Aperture", &exif.Aperture},
		{"FocalLength", &exif.FocalLength},
		{"GPSLatitude", &exif.GPSLatitude},
		{"GPSLongitude", &exif.GPSLongitude},
	}

	for _, tc := range tests {
		t.Run(tc.field, func(t *testing.T) {
			*tc.ptr = &nan
			sanitizeEXIF(&exif)
			if got := *tc.ptr; got != nil {
				t.Errorf("after sanitizeEXIF(), exif.%s = %v, want: nil", tc.field, got)
			}
		})
	}
}

func TestSanitizeEXIF_GPS(t *testing.T) {
	nan := math.NaN()
	valid := float64(10.0)
	var exif models.MediaEXIF

	tests := []struct {
		field string
		ptr   **float64
	}{
		{"GPSLatitude", &exif.GPSLatitude},
		{"GPSLongitude", &exif.GPSLongitude},
	}

	for _, tc := range tests {
		t.Run(tc.field, func(t *testing.T) {
			exif.GPSLatitude = &valid
			exif.GPSLongitude = &valid

			*tc.ptr = &nan
			sanitizeEXIF(&exif)
			if exif.GPSLatitude != nil || exif.GPSLongitude != nil {
				t.Errorf("after sanitizeEXIF(), exif.GPSLatitude = %v, exif.GPSLongitude = %v, want both: nil", exif.GPSLatitude, exif.GPSLongitude)
			}
		})
	}
}
