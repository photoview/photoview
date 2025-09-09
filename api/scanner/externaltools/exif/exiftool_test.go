package exif

import (
	"math"
	"path"
	"testing"
	"time"

	"github.com/barasher/go-exiftool"
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/stretchr/testify/assert"
)

func mustParseTimeInLocal(t *testing.T, str string) time.Time {
	ret, err := time.ParseInLocation(time.DateTime, str, time.Local)
	if err != nil {
		t.Fatalf("invalid time")
	}
	return ret
}

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
				assert.WithinDuration(t, *exif.DateShot, mustParseTimeInLocal(t, "2012-05-06 15:39:44"), time.Minute)
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
	allExif := map[string]string{
		"OffsetTimeOriginal":     "+01:00",
		"OffsetTime":             "+02:00",
		"TimeZone":               "+03:00",
		"SubSecDateTimeOriginal": "2025:09:01 10:00:00.001+04:00",
		"SubSecCreateDate":       "2025:09:01 10:00:00.002+05:00",
		"DateTimeOriginal":       "2025:09:01 10:00:01",
		"MediaCreateDate":        "2025:09:01 08:00:03+06:00",
		"TrackCreateDate":        "2025:09:01 08:00:04+07:00",
		"CreateDate":             "2025:09:01 08:00:05",
		"GPSDateTime":            "2025:09:01 02:00:02Z",
	}

	tests := []struct {
		name            string
		withTags        []string
		wantRFC3339Nano string
		wantErr         bool
	}{
		{"NoTime", []string{}, "", true},

		{
			"SubSecDateTimeOriginal",
			[]string{
				"OffsetTimeOriginal",
				"OffsetTime",
				"TimeZone",
				"SubSecDateTimeOriginal",
				"SubSecCreateDate",
				"DateTimeOriginal",
				"GPSDateTime",
				"MediaCreateDate",
				"TrackCreateDate",
				"CreateDate",
			},
			"2025-09-01T10:00:00.001+04:00",
			false,
		},
		{
			"SubSecCreateDate",
			[]string{
				"OffsetTimeOriginal",
				"OffsetTime",
				"TimeZone",
				"SubSecCreateDate",
				"DateTimeOriginal",
				"GPSDateTime",
				"MediaCreateDate",
				"TrackCreateDate",
				"CreateDate",
			},
			"2025-09-01T10:00:00.002+05:00",
			false,
		},
		{
			"DateTimeOriginal/OffsetTimeOriginal",
			[]string{
				"OffsetTimeOriginal",
				"OffsetTime",
				"TimeZone",
				"DateTimeOriginal",
				"GPSDateTime",
				"MediaCreateDate",
				"TrackCreateDate",
				"CreateDate",
			},
			"2025-09-01T10:00:01+01:00",
			false,
		},
		{
			"DateTimeOriginal/OffsetTime",
			[]string{
				"OffsetTime",
				"TimeZone",
				"DateTimeOriginal",
				"GPSDateTime",
				"MediaCreateDate",
				"TrackCreateDate",
				"CreateDate",
			},
			"2025-09-01T10:00:01+02:00",
			false,
		},
		{
			"DateTimeOriginal/TimeZone",
			[]string{
				"TimeZone",
				"DateTimeOriginal",
				"GPSDateTime",
				"MediaCreateDate",
				"TrackCreateDate",
				"CreateDate",
			},
			"2025-09-01T10:00:01+03:00",
			false,
		},
		{
			"DateTimeOriginal/NoTimezone",
			[]string{
				"DateTimeOriginal",
				"GPSDateTime",
				"MediaCreateDate",
				"TrackCreateDate",
				"CreateDate",
			},
			"2025-09-01T10:00:01+07:59",
			false,
		},
		{
			"GPSDateTime/OffsetTimeOriginal",
			[]string{
				"OffsetTimeOriginal",
				"OffsetTime",
				"TimeZone",
				"GPSDateTime",
				"MediaCreateDate",
				"TrackCreateDate",
				"CreateDate",
			},
			"2025-09-01T03:00:02+01:00",
			false,
		},
		{
			"GPSDateTime/OffsetTime",
			[]string{
				"OffsetTime",
				"TimeZone",
				"GPSDateTime",
				"MediaCreateDate",
				"TrackCreateDate",
				"CreateDate",
			},
			"2025-09-01T04:00:02+02:00",
			false,
		},
		{
			"GPSDateTime/TimeZone",
			[]string{
				"TimeZone",
				"GPSDateTime",
				"MediaCreateDate",
				"TrackCreateDate",
				"CreateDate",
			},
			"2025-09-01T05:00:02+03:00",
			false,
		},
		{
			"GPSDateTime/NoTimezone",
			[]string{
				"GPSDateTime",
				"MediaCreateDate",
				"TrackCreateDate",
				"CreateDate",
			},
			"2025-09-01T02:00:02Z",
			false,
		},
		{
			"MediaCreateDate/OffsetTimeOriginal",
			[]string{
				"OffsetTimeOriginal",
				"OffsetTime",
				"TimeZone",
				"MediaCreateDate",
				"TrackCreateDate",
				"CreateDate",
			},
			"2025-09-01T03:00:03+01:00",
			false,
		},
		{
			"MediaCreateDate/OffsetTime",
			[]string{
				"OffsetTime",
				"TimeZone",
				"MediaCreateDate",
				"TrackCreateDate",
				"CreateDate",
			},
			"2025-09-01T04:00:03+02:00",
			false,
		},
		{
			"MediaCreateDate/TimeZone",
			[]string{
				"TimeZone",
				"MediaCreateDate",
				"TrackCreateDate",
				"CreateDate",
			},
			"2025-09-01T05:00:03+03:00",
			false,
		},
		{
			"MediaCreateDate/NoTimezone",
			[]string{
				"MediaCreateDate",
				"TrackCreateDate",
				"CreateDate",
			},
			"2025-09-01T08:00:03+06:00",
			false,
		},
		{
			"TrackCreateDate/OffsetTimeOriginal",
			[]string{
				"OffsetTimeOriginal",
				"OffsetTime",
				"TimeZone",
				"TrackCreateDate",
				"CreateDate",
			},
			"2025-09-01T02:00:04+01:00",
			false,
		},
		{
			"TrackCreateDate/OffsetTime",
			[]string{
				"OffsetTime",
				"TimeZone",
				"TrackCreateDate",
				"CreateDate",
			},
			"2025-09-01T03:00:04+02:00",
			false,
		},
		{
			"TrackCreateDate/TimeZone",
			[]string{
				"TimeZone",
				"TrackCreateDate",
				"CreateDate",
			},
			"2025-09-01T04:00:04+03:00",
			false,
		},
		{
			"TrackCreateDate/NoTimezone",
			[]string{
				"TrackCreateDate",
				"CreateDate",
			},
			"2025-09-01T08:00:04+07:00",
			false,
		},
		{
			"CreateDate/OffsetTimeOriginal",
			[]string{
				"OffsetTimeOriginal",
				"OffsetTime",
				"TimeZone",
				"CreateDate",
			},
			"2025-09-01T08:00:05+01:00",
			false,
		},
		{
			"CreateDate/OffsetTime",
			[]string{
				"OffsetTime",
				"TimeZone",
				"CreateDate",
			},
			"2025-09-01T08:00:05+02:00",
			false,
		},
		{
			"CreateDate/TimeZone",
			[]string{
				"TimeZone",
				"CreateDate",
			},
			"2025-09-01T08:00:05+03:00",
			false,
		},
		{
			"CreateDate/NoTimezone",
			[]string{
				"CreateDate",
			},
			"2025-09-01T08:00:05" + time.Now().Local().Format("-07:00"),
			false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			metadata := exiftool.EmptyFileMetadata()
			for _, tag := range tc.withTags {
				value, ok := allExif[tag]
				if !ok {
					t.Fatalf("can't get value for exif tag %q", tag)
				}
				metadata.SetString(tag, value)
			}

			gotTime, err := extractDateShot(&metadata)
			gotErr := err != nil
			if gotErr != tc.wantErr {
				t.Fatalf("extractDateShot(%v) returns an error: %v, want an error: %v", tc.withTags, gotErr, tc.wantErr)
			}
			if gotErr {
				return
			}

			if got, want := gotTime.Format(time.RFC3339Nano), tc.wantRFC3339Nano; got != want {
				t.Errorf("extractDateShot(%v) = %v, want: %v", tc.withTags, got, want)
			}
		})
	}
}
