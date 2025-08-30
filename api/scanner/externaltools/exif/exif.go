package exif

import (
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

	log.Info(nil, "Found exiftool")

	return func() {
		if err := globalExifParser.Close(); err != nil {
			log.Error(nil, "Cleanup exiftool error:", err)
			return
		}

		globalExifParser = nil
	}, nil
}

var globalMu sync.Mutex

func Parse(filepath string) (*models.MediaEXIF, error) {
	if globalExifParser == nil {
		return nil, fmt.Errorf("no exif parser initialized")
	}

	globalMu.Lock()
	defer globalMu.Unlock()

	exif, failures, err := globalExifParser.ParseExif(filepath)
	if err != nil {
		return nil, err
	}

	if len(failures) > 0 {
		log.Warn(nil, "Parse exif failures", "filepath", filepath, "errors", failures)
	}

	return exif, nil
}
