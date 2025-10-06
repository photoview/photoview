package exif

import (
	"fmt"
	"math"
	"time"

	"github.com/barasher/go-exiftool"
	"github.com/photoview/photoview/api/graphql/models"
)

// ExifParser is a parser to get exif data.
type ExifParser struct {
	exiftool *exiftool.Exiftool
}

// NewExifParser creates a ExifParser.
func NewExifParser() (*ExifParser, error) {
	buf := make([]byte, 256*1024)

	et, err := exiftool.NewExiftool(exiftool.NoPrintConversion(), exiftool.Buffer(buf, 64*1024))

	if err != nil {
		return nil, fmt.Errorf("error initializing ExifTool: %w", err)
	}

	return &ExifParser{
		exiftool: et,
	}, nil
}

// Close cleans up the buffer of the parser.
func (p *ExifParser) Close() error {
	return p.exiftool.Close()
}

// ParseExif returns the exif data.
func (p *ExifParser) ParseExif(mediaPath string) (*models.MediaEXIF, ParseFailures, error) {
	fileInfo, err := p.readFileInfo(mediaPath)
	if err != nil {
		return nil, nil, err
	}

	retEXIF := models.MediaEXIF{}
	foundExif := false
	var failures ParseFailures

	for field, ptr := range map[string]**string{
		"ImageDescription": &retEXIF.Description,
		"Model":            &retEXIF.Camera, // camera model
		"Make":             &retEXIF.Maker,  // camera make
		"LensModel":        &retEXIF.Lens,
	} {
		value, err := fileInfo.GetString(field)
		if err != nil {
			failures.Append(field, err)
		} else {
			*ptr = &value
			foundExif = true
		}
	}

	for field, ptr := range map[string]**int64{
		"ISO":             &retEXIF.Iso,
		"Flash":           &retEXIF.Flash,
		"Orientation":     &retEXIF.Orientation,
		"ExposureProgram": &retEXIF.ExposureProgram,
	} {
		value, err := fileInfo.GetInt(field)
		if err != nil {
			failures.Append(field, err)
		} else {
			*ptr = &value
			foundExif = true
		}
	}

	for field, ptr := range map[string]**float64{
		"ExposureTime": &retEXIF.Exposure,
		"Aperture":     &retEXIF.Aperture,
		"FocalLength":  &retEXIF.FocalLength,
	} {
		value, err := fileInfo.GetFloat(field)
		if err != nil {
			failures.Append(field, err)
		} else {
			*ptr = &value
			foundExif = true
		}
	}

	// Get time of photo
	date, err := extractDateShot(fileInfo)
	if err != nil {
		failures.Append("DateShot", err)
	} else {
		retEXIF.DateShotStr = &date
		foundExif = true
	}

	// Get GPS data
	lat, long, err := extractValidGPSData(fileInfo)
	if err != nil {
		failures.Append("gps", err)
	} else {
		retEXIF.GPSLatitude, retEXIF.GPSLongitude = &lat, &long
		foundExif = true
	}

	if !foundExif {
		return nil, nil, nil
	}

	sanitizeEXIF(&retEXIF)
	return &retEXIF, failures, nil
}

func (p *ExifParser) readFileInfo(mediaPath string) (*exiftool.FileMetadata, error) {
	fileInfos := p.exiftool.ExtractMetadata(mediaPath)
	if l := len(fileInfos); l != 1 {
		return nil, fmt.Errorf("invalid file infos with %q, len(fileInfos) = %d", mediaPath, l)
	}

	fileInfo := fileInfos[0]
	if err := fileInfo.Err; err != nil {
		return nil, fmt.Errorf("invalid parse %q exif: %w", mediaPath, err)
	}

	return &fileInfo, nil
}

// isFloatReal returns true when the float value represents a real number
// (different than +Inf, -Inf or NaN)
func isFloatReal(v float64) bool {
	if math.IsInf(v, 0) || math.IsNaN(v) {
		return false
	}

	return true
}

// sanitizeEXIF removes any EXIF float64 field that is not a real number (+Inf,
// -Inf or Nan)
func sanitizeEXIF(exif *models.MediaEXIF) {
	if exif.Exposure != nil && !isFloatReal(*exif.Exposure) {
		exif.Exposure = nil
	}

	if exif.Aperture != nil && !isFloatReal(*exif.Aperture) {
		exif.Aperture = nil
	}

	if exif.FocalLength != nil && !isFloatReal(*exif.FocalLength) {
		exif.FocalLength = nil
	}

	if (exif.GPSLatitude != nil && !isFloatReal(*exif.GPSLatitude)) ||
		(exif.GPSLongitude != nil && !isFloatReal(*exif.GPSLongitude)) {
		exif.GPSLatitude = nil
		exif.GPSLongitude = nil
	}
}

func extractValidGPSData(meta *exiftool.FileMetadata) (float64, float64, error) {
	var latitude, longitude *float64

	// GPS coordinates - latitude
	rawLatitude, err := meta.GetFloat("GPSLatitude")
	if err == nil {
		latitude = &rawLatitude
	}

	// GPS coordinates - longitude
	rawLongitude, err := meta.GetFloat("GPSLongitude")
	if err == nil {
		longitude = &rawLongitude
	}

	if latitude == nil || longitude == nil {
		return 0, 0, exiftool.ErrKeyNotFound
	}

	// GPS data validation
	if math.Abs(*latitude) > 90 || math.Abs(*longitude) > 180 {
		latStr := fmt.Sprintf("%f", *latitude)

		longStr := fmt.Sprintf("%f", *longitude)

		return 0, 0, fmt.Errorf("incorrect GPS data: latitude %s should be (-90, 90), longitude %s should be (-180, 180)", latStr, longStr)
	}

	return *latitude, *longitude, nil
}

func extractDateShot(meta *exiftool.FileMetadata) (string, error) {
	var loc *time.Location

TIMEZONE:
	for _, field := range []string{"OffsetTimeOriginal", "OffsetTime", "TimeZone"} {
		str, err := meta.GetString(field)
		if err != nil {
			continue TIMEZONE
		}

		t, err := time.Parse("-07:00", str)
		if err != nil {
			continue TIMEZONE
		}

		_, offsetSecs := t.Zone()
		loc = time.FixedZone(str, offsetSecs)
		break TIMEZONE
	}

	if loc == nil {
		l, err := calculateTimezoneWithUtcGps(meta)
		if err == nil {
			loc = l
		}
	}

	layout := "2006:01:02 15:04:05"
	layoutWithOffset := "2006:01:02 15:04:05Z07:00"
	for _, pattern := range []struct {
		dateField        string
		originalTimezone bool // true if the date in dateField has a timezone when took the image.
	}{
		// Keep the order for the priority to generate DateShot
		{"SubSecDateTimeOriginal", true},
		{"SubSecCreateDate", true},
		{"DateTimeOriginal", false},
		{"GPSDateTime", false},
		{"MediaCreateDate", true},
		{"TrackCreateDate", true},
		{"FileCreateDate", false},
		{"CreateDate", false},
	} {
		dateStr, err := meta.GetString(pattern.dateField)
		if err != nil {
			continue
		}

		if date, err := time.Parse(layoutWithOffset, dateStr); err == nil {
			if !pattern.originalTimezone && loc != nil {
				return date.In(loc).Format(models.RFC3339Milli), nil
			}

			return date.Format(models.RFC3339Milli), nil
		}

		if loc != nil {
			if date, err := time.ParseInLocation(layoutWithOffset, dateStr+loc.String(), loc); err == nil {
				return date.Format(models.RFC3339Milli), nil
			}
		}

		if date, err := time.ParseInLocation(layout, dateStr, time.Local); err == nil {
			return date.Format(models.RFC3339MilliWithoutTimezone), nil
		}
	}

	return "", exiftool.ErrKeyNotFound
}

func calculateTimezoneWithUtcGps(meta *exiftool.FileMetadata) (*time.Location, error) {
	originalStr, err := meta.GetString("DateTimeOriginal")
	if err != nil {
		return nil, err
	}
	original, err := time.Parse("2006:01:02 15:04:05", originalStr)
	if err != nil {
		return nil, err
	}

	gpsStr, err := meta.GetString("GPSDateTime")
	if err != nil {
		return nil, err
	}
	gps, err := time.Parse("2006:01:02 15:04:05Z", gpsStr)
	if err != nil {
		return nil, err
	}

	diff := original.Sub(gps)
	hours := int(diff / time.Hour)
	mins := int(diff/time.Minute) - hours*60
	if mins < 0 {
		mins = -mins
	}
	zoneName := fmt.Sprintf("%+03d:%02d", hours, mins)

	return time.FixedZone(zoneName, int(diff.Seconds())), nil
}
