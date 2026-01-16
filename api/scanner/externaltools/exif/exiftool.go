package exif

import (
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"time"

	"github.com/barasher/go-exiftool"
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/spf13/afero"
)

const layout = "2006:01:02 15:04:05.999"
const layoutWithOffset = "2006:01:02 15:04:05.999-07:00"

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

func (p *ExifParser) loadMetaData(fs afero.Fs, mediaPath string) (*exiftool.FileMetadata, error) {
	// Check if we're using the OS filesystem directly
	_, isOsFs := fs.(*afero.OsFs)

	var pathToRead string
	var cleanup func() error

	if isOsFs {
		// Direct access to the real filesystem
		pathToRead = mediaPath
		cleanup = func() error { return nil }
	} else {
		// Create a temporary file and copy the content from afero.Fs
		tmpFile, err := os.CreateTemp("", "exiftool-*"+filepath.Ext(mediaPath))
		if err != nil {
			return nil, fmt.Errorf("failed to create temporary file: %w", err)
		}
		pathToRead = tmpFile.Name()
		defer tmpFile.Close()

		cleanup = func() error {
			return os.Remove(pathToRead)
		}

		// Copy file content from afero.Fs to temporary file
		srcFile, err := fs.Open(mediaPath)
		if err != nil {
			_ = tmpFile.Close()
			cleanup()
			return nil, fmt.Errorf("failed to open source file: %w", err)
		}
		defer srcFile.Close()

		if _, err := io.Copy(tmpFile, srcFile); err != nil {
			_ = tmpFile.Close()
			cleanup()
			return nil, fmt.Errorf("failed to copy file content: %w", err)
		}

		if err := tmpFile.Close(); err != nil {
			cleanup()
			return nil, fmt.Errorf("failed to close temporary file: %w", err)
		}
	}

	defer cleanup()

	fileInfos := p.exiftool.ExtractMetadata(pathToRead)
	if l := len(fileInfos); l != 1 {
		return nil, fmt.Errorf("expected 1 metadata entry, got %d", l)
	}

	fileInfo := fileInfos[0]
	if err := fileInfo.Err; err != nil {
		return nil, err
	}

	return &fileInfo, nil
}

// ParseExif returns the exif data.
func (p *ExifParser) ParseExif(fs afero.Fs, mediaPath string) (*models.MediaEXIF, ParseFailures, error) {
	fileInfo, err := p.loadMetaData(fs, mediaPath)
	if err != nil {
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
CREATE_DATE:
	for _, createDateKey := range []string{
		// Keep the order for the priority to generate DateShot
		"SubSecDateTimeOriginal",
		"SubSecCreateDate",
		"DateTimeOriginal",
		"CreateDate",
		"TrackCreateDate",
		"MediaCreateDate",
		"FileModifyDate",
	} {
		dateStr, err := fileInfo.GetString(createDateKey)
		if err != nil {
			continue
		}

		if date, err := time.ParseInLocation(layout, dateStr, time.Local); err == nil {
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

	// Get timezone of photo
TIMEZONE:
	for _, field := range []string{
		// Keep the order for the priority to generate TimezoneShot
		"OffsetTimeOriginal",
		"OffsetTime",
		"TimeZone",
	} {
		str, err := fileInfo.GetString(field)
		if err != nil {
			continue TIMEZONE
		}

		t, err := time.Parse("-07:00", str)
		if err != nil {
			failures.Append(field, err)
			continue TIMEZONE
		}

		_, offsetSecs := t.Zone()
		retEXIF.OffsetSecShot = &offsetSecs
		foundExif = true
		break TIMEZONE
	}

	if retEXIF.OffsetSecShot == nil {
		offset, errKeys, err := calculateOffsetFromGPS(fileInfo, retEXIF.DateShot)
		if err != nil {
			for _, key := range errKeys {
				failures.Append(key, err)
			}
		} else {
			retEXIF.OffsetSecShot = offset
		}
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

		return 0, 0, fmt.Errorf(
			"incorrect GPS data: latitude %s should be within [-90, 90], longitude %s should be within [-180, 180]",
			latStr, longStr,
		)
	}

	return *latitude, *longitude, nil
}

func calculateOffsetFromGPS(fileInfo *exiftool.FileMetadata, date *time.Time) (*int, []string, error) {
	if date == nil {
		// There is no original date, can't calculate the offset
		return nil, nil, nil
	}

	const dateKey = "GPSDateStamp"
	dateStr, err := fileInfo.GetString(dateKey)
	if err != nil {
		// Ignore finding-tag errors
		return nil, nil, nil
	}

	const timeKey = "GPSTimeStamp"
	timeStr, err := fileInfo.GetString(timeKey)
	if err != nil {
		// Ignore finding-tag errors
		return nil, nil, nil
	}

	gpsDate, err := time.Parse(layout, dateStr+" "+timeStr)
	if err != nil {
		return nil, []string{dateKey, timeKey}, fmt.Errorf("parse gps date \"%s %s\" error: %w", dateStr, timeStr, err)
	}

	// GPS time is always UTC per EXIF spec
	// offset = GPS UTC time - local time
	offset := int(gpsDate.Sub(*date).Seconds())
	return &offset, nil, nil
}

func (p *ExifParser) ParseMIMEType(fs afero.Fs, mediaPath string) (string, error) {
	fileInfo, err := p.loadMetaData(fs, mediaPath)
	if err != nil {
		return "", fmt.Errorf("invalid parse %q exif: %w", mediaPath, err)
	}

	mime, err := fileInfo.GetString("MIMEType")
	if err != nil {
		if errors.Is(err, exiftool.ErrKeyNotFound) {
			return "", nil // "" is media_type.TypeUnknown
		}
		return "", fmt.Errorf("invalid parse %q mimetype: %w", mediaPath, err)
	}

	return mime, nil
}
