package dataloader

import (
	"time"

	"github.com/barasher/go-exiftool"
)

func NewExiftoolLoader(et *exiftool.Exiftool) *ExiftoolLoader {
	return &ExiftoolLoader{
		wait:     100 * time.Millisecond,
		maxBatch: 100,
		fetch: func(keys []string) ([]exiftool.FileMetadata, []error) {
			metadata := et.ExtractMetadata(keys...)

			exifErrors := make([]error, len(metadata))
			for i := 0; i < len(metadata); i++ {
				exifErrors[i] = metadata[i].Err
			}

			return metadata, exifErrors
		},
	}
}
