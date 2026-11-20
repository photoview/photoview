package scanner_tasks

import (
	"context"
	"testing"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner/face_detection"
	"github.com/photoview/photoview/api/scanner/media_encoding"
	"github.com/photoview/photoview/api/scanner/scanner_task"
	"github.com/photoview/photoview/api/test_utils"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestMain(m *testing.M) {
	test_utils.IntegrationTestRun(m)
}

// recordingDetector counts scan detection separately from the user-triggered
// recognition mutation. The authorization filter applies only to that mutation.
type recordingDetector struct {
	detectCalls    int
	recognizeCalls int
}

func (*recordingDetector) ReloadFacesFromDatabase(*gorm.DB) error { return nil }
func (d *recordingDetector) DetectFaces(*gorm.DB, *models.Media) error {
	d.detectCalls++
	return nil
}
func (*recordingDetector) MergeCategories(int32, int32) {}
func (*recordingDetector) MergeImageFaces([]int, int32) {}
func (d *recordingDetector) RecognizeUnlabeledFaces(*gorm.DB, *models.User) ([]*models.ImageFace, error) {
	d.recognizeCalls++
	return nil, nil
}

func TestFaceScanDetectsFacesWithoutRecognizingUnlabeledResults(t *testing.T) {
	db := test_utils.DatabaseTest(t)

	previous := face_detection.GlobalFaceDetector
	detector := &recordingDetector{}
	face_detection.GlobalFaceDetector = detector
	t.Cleanup(func() { face_detection.GlobalFaceDetector = previous })

	photo := &models.Media{Type: models.MediaTypePhoto, Path: "/photos/a.jpg"}
	data := &media_encoding.EncodeMediaData{Media: photo}
	ctx := scanner_task.NewTaskContext(context.Background(), db, &models.Album{}, nil)

	err := (FaceDetectionTask{}).AfterProcessMedia(ctx, data, []*models.MediaURL{{}}, 0, 1)
	require.NoError(t, err)
	require.Equal(t, 1, detector.detectCalls)
	require.Equal(t, 0, detector.recognizeCalls)

	video := &models.Media{Type: models.MediaTypeVideo, Path: "/photos/a.mp4"}
	videoData := &media_encoding.EncodeMediaData{Media: video}
	err = (FaceDetectionTask{}).AfterProcessMedia(ctx, videoData, []*models.MediaURL{{}}, 0, 1)
	require.NoError(t, err)
	require.Equal(t, 1, detector.detectCalls)
	require.Equal(t, 0, detector.recognizeCalls)

	err = (FaceDetectionTask{}).AfterProcessMedia(ctx, data, nil, 0, 1)
	require.NoError(t, err)
	require.Equal(t, 1, detector.detectCalls, "a photo that was not re-encoded is not scanned again")
	require.Equal(t, 0, detector.recognizeCalls)
}
