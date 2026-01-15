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

func (t FaceDetectionTask) AfterProcessMedia(ctx scanner_task.TaskContext, mediaData *media_encoding.EncodeMediaData,
	updatedURLs []*models.MediaURL, mediaIndex int, mediaTotal int) error {

	didProcess := len(updatedURLs) > 0

	if didProcess && mediaData.Media.Type == models.MediaTypePhoto {
		media := mediaData.Media
		if face_detection.GlobalFaceDetector == nil {
			return nil
		}
		if err := face_detection.GlobalFaceDetector.DetectFaces(ctx.GetDB(), ctx.GetFS(), media); err != nil {
			scanner_utils.ScannerError(ctx, "Error detecting faces in image (%s): %s", media.Path, err)
		}
	}

	return nil
}
