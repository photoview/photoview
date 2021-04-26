package scanner_test

import (
	"os"
	"testing"
	"time"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner"
	"github.com/photoview/photoview/api/scanner/face_detection"
	"github.com/photoview/photoview/api/test_utils"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	os.Exit(test_utils.IntegrationTestRun(m))
}

func TestFullScan(t *testing.T) {
	test_utils.FilesystemTest(t)
	db := test_utils.DatabaseTest(t)

	pass := "1234"
	user, err := models.RegisterUser(db, "test_user", &pass, true)
	if !assert.NoError(t, err) {
		return
	}

	root_album := models.Album{
		Title: "root album",
		Path:  "./test_data",
	}

	if !assert.NoError(t, db.Save(&root_album).Error) {
		return
	}

	err = db.Model(user).Association("Albums").Append(&root_album)
	if !assert.NoError(t, err) {
		return
	}

	if !assert.NoError(t, scanner.InitializeScannerQueue(db)) {
		return
	}

	if !assert.NoError(t, face_detection.InitializeFaceDetector(db)) {
		return
	}

	if !assert.NoError(t, scanner.AddUserToQueue(user)) {
		return
	}

	// wait for all jobs to finish
	scanner.CloseScannerQueue()

	var all_media []*models.Media
	if !assert.NoError(t, db.Find(&all_media).Error) {
		return
	}

	assert.Equal(t, 9, len(all_media))

	var all_media_url []*models.MediaURL
	if !assert.NoError(t, db.Find(&all_media_url).Error) {
		return
	}

	assert.Equal(t, 18, len(all_media_url))

	// Verify that faces was recognized
	assert.Eventually(t, func() bool {
		var all_face_groups []*models.FaceGroup
		if !assert.NoError(t, db.Find(&all_face_groups).Error) {
			return false
		}

		return len(all_face_groups) == 3
	}, time.Second*5, time.Millisecond*500)

	assert.Eventually(t, func() bool {
		var all_image_faces []*models.ImageFace
		if !assert.NoError(t, db.Find(&all_image_faces).Error) {
			return false
		}

		return len(all_image_faces) == 6
	}, time.Second*5, time.Millisecond*500)

}
