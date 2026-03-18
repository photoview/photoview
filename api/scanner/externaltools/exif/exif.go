package exif

import (
	"fmt"
	"sync"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/log"
	"github.com/photoview/photoview/api/scanner/externaltools/exiftool"
)

var globalExifParser *exiftool.Exiftool
var globalInit sync.Once

func Initialize() (func(), error) {
	var err error
	globalInit.Do(func() {
		globalExifParser, err = exiftool.New()
	})

	if err != nil {
		return nil, err
	}

	log.Info(nil, "Found exiftool.", "binary_path", globalExifParser.BinaryPath(), "version", globalExifParser.Version())

	return func() {
		globalMu.Lock()
		defer globalMu.Unlock()

		if globalExifParser == nil {
			return
		}

		if err := globalExifParser.Close(); err != nil {
			log.Error(nil, "Cleanup exiftool error", "error", err)
			return
		}
		globalExifParser = nil
	}, nil
}

var globalMu sync.Mutex

func Parse(filepath string) (*models.MediaEXIF, error) {
	globalMu.Lock()
	defer globalMu.Unlock()

	if globalExifParser == nil {
		return nil, fmt.Errorf("no exif parser initialized")
	}

	var values struct {
		exiftool.PhotoMeta
		exiftool.TimeAll
		exiftool.GPS
	}
	if err := globalExifParser.QueryJSONTagsByNumber(filepath, &values); err != nil {
		return nil, err
	}

	values.PhotoMeta.SanitizeFloats()

	ret := models.MediaEXIF{
		Camera:          values.Model,
		Maker:           values.Make,
		Lens:            values.LensModel,
		Iso:             values.ISO,
		Flash:           values.Flash,
		Orientation:     values.Orientation,
		ExposureProgram: values.ExposureProgram,
		Exposure:        values.ExposureTime,
		Aperture:        values.Aperture,
		FocalLength:     values.FocalLength,
		Description:     values.ImageDescription,
	}

	dateShot := values.TimeAll.TimeInLocal()
	if !dateShot.IsZero() {
		ret.DateShot = new(dateShot)
	}

	offsetSec, ok := values.TimeAll.OffsetSecs(dateShot)
	if ok {
		ret.OffsetSecShot = &offsetSec
	}

	if values.GPS.IsValid() {
		ret.GPSLatitude = values.GPS.GPSLatitude
		ret.GPSLongitude = values.GPS.GPSLongitude
	}

	return &ret, nil
}

func MIMEType(filepath string) (string, error) {
	globalMu.Lock()
	defer globalMu.Unlock()

	if globalExifParser == nil {
		return "", fmt.Errorf("no exif parser initialized")
	}

	var mime exiftool.MIMEType
	if err := globalExifParser.QueryJSONTagsByNumber(filepath, &mime); err != nil {
		return "", err
	}

	if mime.MIMEType == nil {
		return "", nil
	}

	return *mime.MIMEType, nil

}
