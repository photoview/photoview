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
	fileInfos := p.exiftool.ExtractMetadata(mediaPath)
	if l := len(fileInfos); l != 1 {
		return nil, nil, fmt.Errorf("invalid file infos with %q, len(fileInfos) = %d", mediaPath, l)
	}

	fileInfo := fileInfos[0]
	if err := fileInfo.Err; err != nil {
		return nil, nil, fmt.Errorf("invalid parse %q exif: %w", mediaPath, err)
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
	layout := "2006:01:02 15:04:05"
	layoutWithOffset := "2006:01:02 15:04:05-07:00"
CREATE_DATE:
	for _, createDateKey := range []string{
		// Keep the order for the priority to generate DateShot
		"CreationDate",
		"DateTimeOriginal",
		"CreateDate",
		"TrackCreateDate",
		"MediaCreateDate",
		"FileCreateDate",
		"ModifyDate",
		"TrackModifyDate",
		"MediaModifyDate",
		"FileModifyDate",
	} {
		dateStr, err := fileInfo.GetString(createDateKey)
		if err != nil {
			continue
		}

		if date, err := time.Parse(layout, dateStr); err == nil {
			retEXIF.DateShot = &date
			foundExif = true
			break CREATE_DATE
		}

		if date, err := time.Parse(layoutWithOffset, dateStr); err == nil {
			retEXIF.DateShot = &date
			foundExif = true
			break CREATE_DATE
		} else {
			failures.Append(createDateKey, err)
		}
	}

	// Get GPS data
	lat, long, err := extractValidGPSData(&fileInfo)
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

func extractValidGPSData(fileInfo *exiftool.FileMetadata) (float64, float64, error) {
	var latitude, longitude *float64

	// GPS coordinates - latitude
	rawLatitude, err := fileInfo.GetFloat("GPSLatitude")
	if err == nil {
		latitude = &rawLatitude
	}

	// GPS coordinates - longitude
	rawLongitude, err := fileInfo.GetFloat("GPSLongitude")
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
