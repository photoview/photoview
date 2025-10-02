package exif

import (
	"context"
	"fmt"
	"sync"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/log"
)

var globalExifParser *ExifParser
var globalInit sync.Once

func Initialize() (func(), error) {
	var err error
	globalInit.Do(func() {
		globalExifParser, err = NewExifParser()
	})

	if err != nil {
		return nil, err
	}

	log.Info(context.Background(), "Found exiftool")

	return func() {
		globalMu.Lock()
		defer globalMu.Unlock()

		if globalExifParser == nil {
			return
		}

		if err := globalExifParser.Close(); err != nil {
			log.Error(context.Background(), "Cleanup exiftool error", "error", err)
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

	exif, failures, err := globalExifParser.ParseExif(filepath)
	if err != nil {
		return nil, err
	}

	if len(failures) > 0 {
		log.Warn(context.Background(), "Parse exif failures", "filepath", filepath, "errors", failures)
	}

	return exif, nil
}
