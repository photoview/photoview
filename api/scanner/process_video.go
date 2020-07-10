package scanner

import (
	"database/sql"
	"log"

	"github.com/pkg/errors"
	"github.com/viktorstrate/photoview/api/graphql/models"
)

func processVideo(tx *sql.Tx, imageData *EncodeImageData, videoCachePath *string) (bool, error) {
	video := imageData.media
	didProcess := false

	log.Printf("Processing video: %s", video.Path)

	mediaUrlFromDB, err := makePhotoURLChecker(tx, video.MediaID)
	if err != nil {
		return false, err
	}

	videoWebURL, err := mediaUrlFromDB(models.VideoWeb)
	if err != nil {
		return false, errors.Wrap(err, "error processing video web-format")
	}

	if videoWebURL == nil {
		// TODO: Process web video
	}

	// TODO: Process video thumbnail

	return didProcess, nil
}
