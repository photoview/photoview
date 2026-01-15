package exif

import (
	"math"
	"path"
	"testing"
	"time"

	"github.com/barasher/go-exiftool"
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func parseRFC3339(t *testing.T, str string) time.Time {
	t.Helper()

	ret, err := time.Parse(time.RFC3339Nano, str)
	if err != nil {
		t.Fatalf("invalid time %q: %v", str, err)
	}
	return ret
}

func TestExifParser(t *testing.T) {
	fs := afero.NewOsFs()

	parser, err := NewExifParser()
	if err != nil {
		t.Fatalf("can't init exiftool: %v", err)
	}
	defer parser.Close()

	images := []struct {
		path   string
		assert func(t *testing.T, exif *models.MediaEXIF, err error)
	}{
		{
			path: "./test_data/bird.jpg",
			assert: func(t *testing.T, exif *models.MediaEXIF, err error) {
				assert.NoError(t, err)
				assert.EqualValues(t, *exif.Description, "Photo of a Bird")
				assert.EqualValues(t, *exif.Camera, "Canon EOS 600D")
				assert.EqualValues(t, *exif.Maker, "Canon")
				assert.WithinDuration(t, *exif.DateShot, time.Unix(1336318784, 0), time.Minute)
				assert.Nil(t, exif.OffsetSecShot, "OffsetSecShot should be calculated from GPS data")
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
		{
			// stripped.jpg has a file modified date with the offset.
			path: "./test_data/stripped.jpg",
			assert: func(t *testing.T, exif *models.MediaEXIF, err error) {
				assert.NoError(t, err)
				assert.Equal(t, 0, exif.ID)
				assert.True(t, exif.CreatedAt.IsZero())
				assert.True(t, exif.UpdatedAt.IsZero())
				assert.Nil(t, exif.Description)
				assert.Nil(t, exif.Camera)
				assert.Nil(t, exif.Maker)
				assert.Nil(t, exif.OffsetSecShot)
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
			// exif_subsec_timezone.heic has a file with the offset and sub sec.
			path: "./test_data/exif_subsec_timezone.heic",
			assert: func(t *testing.T, exif *models.MediaEXIF, err error) {
				assert.NoError(t, err)
				assert.Equal(t, 0, exif.ID)
				assert.True(t, exif.CreatedAt.IsZero())
				assert.True(t, exif.UpdatedAt.IsZero())
				assert.Nil(t, exif.Description)
				assert.EqualValues(t, *exif.Camera, "iPhone 15 Pro")
				assert.EqualValues(t, *exif.Maker, "Apple")
				assert.WithinDuration(t, *exif.DateShot, parseRFC3339(t, "2025-10-28T14:20:22.164+01:00"), time.Millisecond)
				assert.Equal(t, *exif.OffsetSecShot, 3600)
				assert.Equal(t, *exif.Lens, "iPhone 15 Pro back triple camera 2.22mm f/2.2")
				assert.InDelta(t, *exif.Exposure, 0.05, 0.0001)
				assert.EqualValues(t, *exif.Aperture, 2.2)
				assert.EqualValues(t, *exif.Iso, 1250)
				assert.InDelta(t, *exif.FocalLength, 2.22, 0.005)
				assert.EqualValues(t, *exif.Flash, 16)
				assert.EqualValues(t, *exif.Orientation, 6)
				assert.EqualValues(t, *exif.ExposureProgram, 2)
				assert.Nil(t, exif.GPSLatitude)
				assert.Nil(t, exif.GPSLongitude)
			},
		},
	}

	for _, img := range images {
		t.Run(path.Base(img.path), func(t *testing.T) {
			exif, failures, err := parser.ParseExif(fs, img.path)
			if len(failures) != 0 {
				t.Errorf("parse failures: %v", failures)
			}

			img.assert(t, exif, err)
		})
	}
}

func TestExifParserWithFailure(t *testing.T) {
	fs := afero.NewOsFs()

	parser, err := NewExifParser()
	if err != nil {
		t.Fatalf("can't init exiftool: %v", err)
	}
	defer parser.Close()

	imagesWithFailures := []struct {
		path   string
		assert func(t *testing.T, exif *models.MediaEXIF, err error)
	}{
		{
			path: "./test_data/IncorrectGPS.jpg",
			assert: func(t *testing.T, exif *models.MediaEXIF, err error) {
				assert.Nil(t, exif.GPSLatitude,
					"GPSLatitude expected to be NULL for an incorrect input data: %+v", exif.GPSLatitude)
				assert.Nil(t, exif.GPSLongitude,
					"GPSLongitude expected to be NULL for an incorrect input data: %+v", exif.GPSLongitude)
			},
		},
	}

	for _, img := range imagesWithFailures {
		t.Run(path.Base(img.path), func(t *testing.T) {
			exif, failures, err := parser.ParseExif(fs, img.path)
			if len(failures) == 0 {
				t.Errorf("parse failures: %v, should have at least one failure", failures)
			}

			img.assert(t, exif, err)
		})
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

			if got, want := lat, tc.latitude; math.Abs(got-want) >= math.SmallestNonzeroFloat64 {
				t.Fatalf("extractValidGPSData({lat: %f, long: %f}) got latitude: %v, want: %v", tc.latitude, tc.longitude, got, want)
			}
			if got, want := long, tc.longitude; math.Abs(got-want) >= math.SmallestNonzeroFloat64 {
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
		{"NaN", math.NaN(), false},
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

func parseTime(t *testing.T, str string) *time.Time {
	t.Helper()

	ret, err := time.Parse(layout, str)
	if err == nil {
		return &ret
	}

	ret, err = time.Parse(layoutWithOffset, str)
	if err == nil {
		return &ret
	}

	t.Fatalf("invalid time %q: %v", str, err)
	panic(0)
}

func p[T any](v T) *T {
	return &v
}

func TestCalculateOffsetFromGPS(t *testing.T) {
	tests := []struct {
		name         string
		originalDate *time.Time
		gpsDateStamp *string
		gpsTimeStamp *string
		wantOffset   *int
		wantError    bool
	}{
		{"UTC+1", parseTime(t, "2025:11:04 14:00:00"), p("2025:11:04"), p("15:00:00"), p(60 * 60), false},
		{"UTC-1", parseTime(t, "2025:11:04 14:00:00"), p("2025:11:04"), p("13:00:00"), p(-1 * 60 * 60), false},
		{"UTC+1:15", parseTime(t, "2025:11:04 14:00:00"), p("2025:11:04"), p("15:15:00"), p(60*60 + 15*60), false},
		{"UTC-1:15", parseTime(t, "2025:11:04 14:00:00"), p("2025:11:04"), p("12:45:00"), p(-1*60*60 - 15*60), false},
		{"UTC+8", parseTime(t, "2025:11:04 23:00:00"), p("2025:11:05"), p("07:00:00"), p(8 * 60 * 60), false},
		{"UTC-8", parseTime(t, "2025:11:04 01:00:00"), p("2025:11:03"), p("17:00:00"), p(-8 * 60 * 60), false},
		{"NoOriginalDate", nil, p("2025:11:03"), p("17:00:00"), nil, false},
		{"NoGPSDateStamp", parseTime(t, "2025:11:04 14:00:00"), nil, p("15:00:00"), nil, false},
		{"NoGPSTimeStamp", parseTime(t, "2025:11:04 14:00:00"), p("2025:11:04"), nil, nil, false},
		{"NoGPS", parseTime(t, "2025:11:04 14:00:00"), nil, nil, nil, false},
		{"NoData", nil, nil, nil, nil, false},

		{"InvalidGPSDateStamp", parseTime(t, "2025:11:04 14:00:00"), p("0000:00:00"), p("15:00:00"), nil, true},
		{"InvalidGPSTimeStamp", parseTime(t, "2025:11:04 14:00:00"), p("2025:11:04"), p("25:00:00"), nil, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fileInfo := exiftool.EmptyFileMetadata()

			if tc.gpsDateStamp != nil {
				fileInfo.SetString("GPSDateStamp", *tc.gpsDateStamp)
			}
			if tc.gpsTimeStamp != nil {
				fileInfo.SetString("GPSTimeStamp", *tc.gpsTimeStamp)
			}

			gotOffset, _, err := calculateOffsetFromGPS(&fileInfo, tc.originalDate)
			if gotErr := err != nil; gotErr != tc.wantError {
				t.Errorf("got error: %v, want error: %v", err, tc.wantError)
			}
			if err != nil || tc.wantError {
				return
			}

			if gotOffset == nil && tc.wantOffset == nil {
				return
			}
			if gotOffset == nil || tc.wantOffset == nil {
				t.Errorf("got offset: %+v, want offset: %+v", gotOffset, tc.wantOffset)
				return
			}
			if got, want := *gotOffset, *tc.wantOffset; got != want {
				t.Errorf("got offset: %d, want offset: %d", got, want)
			}
		})
	}
}

func TestParseMIMEType(t *testing.T) {
	fs := afero.NewOsFs()

	parser, err := NewExifParser()
	if err != nil {
		t.Fatalf("can't init exiftool: %v", err)
	}
	defer parser.Close()

	tests := []struct {
		file string
		want string
	}{
		{"./test_data/bird.jpg", "image/jpeg"},
		{"./test_data/exif_subsec_timezone.heic", "image/heic"},
	}

	for _, tc := range tests {
		t.Run(tc.file, func(t *testing.T) {
			got, err := parser.ParseMIMEType(fs, tc.file)
			if err != nil {
				t.Fatalf("ParseMIMEType(%q) returns an error: %v", tc.file, err)
			}

			if got != tc.want {
				t.Errorf("ParseMIMEType(%q) = %q, want: %q", tc.file, got, tc.want)
			}
		})
	}
}
