package exif

import (
	"math"
	"path"
	"testing"

	"github.com/barasher/go-exiftool"
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/stretchr/testify/assert"
)

func TestExifParser(t *testing.T) {
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
				assert.EqualValues(t, *exif.DateShotStr, "2012-05-06T15:39:44.000")
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
				assert.Nil(t, exif)
			},
		},
		{
			path: "./test_data/bad-exif.jpg",
			assert: func(t *testing.T, exif *models.MediaEXIF, err error) {
				assert.NoError(t, err)
				assert.Nil(t, exif.Exposure)
			},
		},
	}

	for _, img := range images {
		t.Run(path.Base(img.path), func(t *testing.T) {
			exif, failures, err := parser.ParseExif(img.path)
			if len(failures) != 0 {
				t.Errorf("parse failures: %v", failures)
			}

			img.assert(t, exif, err)
		})
	}
}

func TestExifParserWithFailure(t *testing.T) {
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
			exif, failures, err := parser.ParseExif(img.path)
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

func TestExtractDateShot(t *testing.T) {
	tests := []struct {
		name     string
		withTags map[string]string
		want     string
		wantErr  bool
	}{
		{"NoTime", nil, "", true},

		{
			"SubSecDateTimeOriginal",
			map[string]string{
				"OffsetTimeOriginal":     "+01:00",
				"SubSecDateTimeOriginal": "2025:01:01 10:00:00.001+04:00",
				"SubSecCreateDate":       "2025:01:01 10:00:00.002+05:00",
			},
			"2025-01-01T10:00:00.001+04:00",
			false,
		},
		{
			"SubSecCreateDate",
			map[string]string{
				"OffsetTimeOriginal": "-01:00",
				"SubSecCreateDate":   "2025:01:31 10:00:00.002-05:00",
				"DateTimeOriginal":   "2025:01:31 10:00:01",
			},
			"2025-01-31T10:00:00.002-05:00",
			false,
		},
		{
			"DateTimeOriginalWithOffsetTimeOriginal",
			map[string]string{
				"OffsetTimeOriginal": "+01:30",
				"OffsetTime":         "+02:30",
				"DateTimeOriginal":   "2025:04:01 01:00:01",
				"MediaCreateDate":    "2025:04:01 01:00:03+06:30",
			},
			"2025-04-01T01:00:01.000+01:30",
			false,
		},
		{
			"DateTimeOriginalWithOffsetTime",
			map[string]string{
				"OffsetTime":       "+02:15",
				"TimeZone":         "+03:15",
				"DateTimeOriginal": "2025:04:30 23:00:01",
				"MediaCreateDate":  "2025:04:30 23:00:03+06:15",
			},
			"2025-04-30T23:00:01.000+02:15",
			false,
		},
		{
			"DateTimeOriginalWithTimeZone",
			map[string]string{
				"TimeZone":         "+03:00",
				"DateTimeOriginal": "2025:06:01 23:00:01",
				"MediaCreateDate":  "2025:06:01 23:00:03+06:15",
				"GPSDateTime":      "2025:06:01 22:00:02Z",
			},
			"2025-06-01T23:00:01.000+03:00",
			false,
		},
		{
			"DateTimeOriginalWithGPSTime",
			map[string]string{
				"DateTimeOriginal": "2025:06:14 23:00:01",
				"MediaCreateDate":  "2025:06:14 23:00:03+06:15",
				"GPSDateTime":      "2025:06:15 00:00:01Z",
			},
			"2025-06-14T23:00:01.000-01:00",
			false,
		},
		{
			"DateTimeOriginalNoTimezone",
			map[string]string{
				"DateTimeOriginal": "2025:06:30 23:59:59",
				"MediaCreateDate":  "2025:06:30 23:59:59+06:15",
			},
			"2025-06-30T23:59:59.000",
			false,
		},
		{
			"GPSDateTimeWithOffsetTimeOriginal",
			map[string]string{
				"OffsetTimeOriginal": "+01:00",
				"OffsetTime":         "+02:00",
				"GPSDateTime":        "2025:11:01 02:00:00Z",
				"MediaCreateDate":    "2025:11:01 01:00:00+06:00",
			},
			"2025-11-01T03:00:00.000+01:00",
			false,
		},
		{
			"GPSDateTimeWithOffsetTime",
			map[string]string{
				"OffsetTime":      "+02:00",
				"TimeZone":        "+03:00",
				"GPSDateTime":     "2025:11:01 02:00:00Z",
				"MediaCreateDate": "2025:11:01 23:00:03+06:00",
			},
			"2025-11-01T04:00:00.000+02:00",
			false,
		},
		{
			"GPSDateTimeWithTimeZone",
			map[string]string{
				"TimeZone":        "+03:00",
				"GPSDateTime":     "2025:11:01 02:00:00Z",
				"MediaCreateDate": "2025:11:01 02:00:00+06:00",
			},
			"2025-11-01T05:00:00.000+03:00",
			false,
		},
		{
			"DateTimeOriginalNoTimezone",
			map[string]string{
				"GPSDateTime":     "2025:11:01 02:00:00Z",
				"MediaCreateDate": "2025:11:01 02:00:00+06:00",
			},
			"2025-11-01T02:00:00.000Z",
			false,
		},
		{
			"MediaCreateDate",
			map[string]string{
				"OffsetTimeOriginal": "+01:00",
				"MediaCreateDate":    "2025:07:01 23:59:59+06:15",
				"TrackCreateDate":    "2025:07:01 23:00:00+06:15",
			},
			"2025-07-01T23:59:59.000+06:15",
			false,
		},
		{
			"TrackCreateDate",
			map[string]string{
				"OffsetTimeOriginal": "-01:00",
				"TrackCreateDate":    "2025:07:31 23:00:00-06:15",
				"CreateDate":         "2025:07:31 23:00:05",
			},
			"2025-07-31T23:00:00.000-06:15",
			false,
		},
		{
			"CreateDateWithOffsetTimeOriginal",
			map[string]string{
				"OffsetTimeOriginal": "+01:30",
				"OffsetTime":         "+02:30",
				"CreateDate":         "2025:12:01 01:00:01",
			},
			"2025-12-01T01:00:01.000+01:30",
			false,
		},
		{
			"CreateDateWithOffsetTime",
			map[string]string{
				"OffsetTime": "+02:15",
				"TimeZone":   "+03:15",
				"CreateDate": "2025:12:01 01:00:01",
			},
			"2025-12-01T01:00:01.000+02:15",
			false,
		},
		{
			"CreateDateWithTimeZone",
			map[string]string{
				"TimeZone":   "+03:00",
				"CreateDate": "2025:12:01 01:00:01",
			},
			"2025-12-01T01:00:01.000+03:00",
			false,
		},
		{
			"CreateDateNoTimezone",
			map[string]string{
				"CreateDate": "2025:12:01 01:00:01",
			},
			"2025-12-01T01:00:01.000",
			false,
		},

		{
			"45MinuteOffset",
			map[string]string{
				"DateTimeOriginal":   "2023:03:15 10:00:00",
				"OffsetTimeOriginal": "+05:45",
			},
			"2023-03-15T10:00:00.000+05:45",
			false,
		},
		{
			"LargePositiveOffset",
			map[string]string{
				"CreateDate": "2024:02:29 23:59:59",
				"OffsetTime": "+14:00",
			},
			"2024-02-29T23:59:59.000+14:00",
			false,
		},
		{
			"YearTransition",
			map[string]string{
				"DateTimeOriginal":   "2024:12:31 23:59:59",
				"OffsetTimeOriginal": "-05:00",
			},
			"2024-12-31T23:59:59.000-05:00",
			false,
		},

		{
			"InvalidMonth",
			map[string]string{
				"DateTimeOriginal":   "2025:13:00 00:00:00",
				"OffsetTimeOriginal": "-05:00",
			},
			"",
			true,
		},
		{
			"InvalidDay",
			map[string]string{
				"DateTimeOriginal":   "2025:12:32 00:00:00",
				"OffsetTimeOriginal": "-05:00",
			},
			"",
			true,
		},
		{
			"InvalidHour",
			map[string]string{
				"DateTimeOriginal":   "2025:12:31 24:00:00",
				"OffsetTimeOriginal": "-05:00",
			},
			"",
			true,
		},
		{
			"InvalidMinute",
			map[string]string{
				"DateTimeOriginal":   "2025:12:31 23:60:00",
				"OffsetTimeOriginal": "-05:00",
			},
			"",
			true,
		},
		{
			"InvalidSecond",
			map[string]string{
				"DateTimeOriginal":   "2025:12:31 23:00:60",
				"OffsetTimeOriginal": "-05:00",
			},
			"",
			true,
		},
		{
			"InvalidTimezone",
			map[string]string{
				"DateTimeOriginal":   "2025:12:31 23:00:00",
				"OffsetTimeOriginal": "-0a:00",
			},
			"2025-12-31T23:00:00.000",
			false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			metadata := exiftool.EmptyFileMetadata()
			for tag, value := range tc.withTags {
				metadata.SetString(tag, value)
			}

			got, err := extractDateShot(&metadata)
			gotErr := err != nil
			if gotErr != tc.wantErr {
				t.Fatalf("extractDateShot(%v) returns an error: %v, want an error: %v", tc.withTags, gotErr, tc.wantErr)
			}
			if gotErr {
				return
			}

			if got, want := got, tc.want; got != want {
				t.Errorf("extractDateShot(%v) = %v, want: %v", tc.withTags, got, want)
			}
		})
	}
}
