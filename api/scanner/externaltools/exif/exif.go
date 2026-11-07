package exif

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

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

// queryTimeout bounds one shared exiftool request. The scanner holds globalMu
// until it returns, so a stuck process can stall every worker for at most this long.
var queryTimeout = time.Minute

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
	if err := queryJSON(filepath, &values); err != nil {
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
	if err := queryJSON(filepath, &mime); err != nil {
		return "", err
	}

	if mime.MIMEType == nil {
		return "", nil
	}

	return *mime.MIMEType, nil

}

// queryJSON runs one bounded exiftool query. Caller holds globalMu.
// A timed-out parser is replaced before the lock is released.
func queryJSON(file string, value any) error {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	err := globalExifParser.QueryJSONTagsByNumberContext(ctx, file, value)
	if err == nil || !isQueryTimeout(err) {
		return err
	}

	log.Warn(nil, "exiftool query timed out; restarting parser", "file", file, "error", err)

	if restartErr := restartExifParser(); restartErr != nil {
		return fmt.Errorf("%w; %w", err, restartErr)
	}

	return err
}

func isQueryTimeout(err error) bool {
	return errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled)
}

// restartExifParser replaces a force-terminated parser. Caller holds globalMu.
func restartExifParser() error {
	old := globalExifParser
	globalExifParser = nil

	if old != nil {
		if err := old.Close(); err != nil {
			log.Error(nil, "Cleanup exiftool error", "error", err)
		}
	}

	parser, err := exiftool.New()
	if err != nil {
		return fmt.Errorf("restart exiftool: %w", err)
	}

	globalExifParser = parser
	log.Info(nil, "Restarted exiftool.", "binary_path", parser.BinaryPath(), "version", parser.Version())

	return nil
}
