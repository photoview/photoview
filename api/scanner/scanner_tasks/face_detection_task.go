package scanner_tasks

import (
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner/face_detection"
	"github.com/photoview/photoview/api/scanner/media_encoding"
	"github.com/photoview/photoview/api/scanner/scanner_task"
	"github.com/photoview/photoview/api/scanner/scanner_utils"
)

type FaceDetectionTask struct {
	scanner_task.ScannerTaskBase
}

func (t FaceDetectionTask) AfterProcessMedia(ctx scanner_task.TaskContext, mediaData *media_encoding.EncodeMediaData, updatedURLs []*models.MediaURL, mediaIndex int, mediaTotal int) error {
	didProcess := len(updatedURLs) > 0

	if didProcess && mediaData.Media.Type == models.MediaTypePhoto {
		go func(media *models.Media) {
			if face_detection.GlobalFaceDetector == nil {
				return
			}
			if err := face_detection.GlobalFaceDetector.DetectFaces(ctx.GetDB(), media, false); err != nil {
				scanner_utils.ScannerError("Error detecting faces in image (%s): %s", media.Path, err)
			}
		}(mediaData.Media)
	}

	return nil
}
