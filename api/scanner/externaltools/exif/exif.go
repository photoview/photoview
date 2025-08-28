package exif

import (
	"fmt"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/log"
)

var globalExifParser *ExifParser

func Initialize() (func(), error) {
	var err error
	globalExifParser, err = NewExifParser()
	if err != nil {
		return nil, err
	}

	log.Info(nil, "Found exiftool")

	return func() {
		if err := globalExifParser.Close(); err != nil {
			log.Error(nil, "Cleanup exiftool error:", err)
		}
	}, nil
}

func Parse(filepath string) (*models.MediaEXIF, error) {
	if globalExifParser == nil {
		return nil, fmt.Errorf("no exif parser initialized")
	}

	return globalExifParser.ParseExif(filepath)
}
