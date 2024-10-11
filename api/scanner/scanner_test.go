package scanner_test

import (
	"os"
	"testing"
	"time"

	"github.com/photoview/photoview/api/graphql/models"
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

	rootAlbum := models.Album{
		Title: "root album",
		Path:  "./test_data",
	}

	if !assert.NoError(t, db.Save(&rootAlbum).Error) {
		return
	}

	err = db.Model(user).Association("Albums").Append(&rootAlbum)
	if !assert.NoError(t, err) {
		return
	}

	if !assert.NoError(t, face_detection.InitializeFaceDetector(db)) {
		return
	}

	test_utils.RunScannerOnUser(t, db, user)

	var allMedia []*models.Media
	if !assert.NoError(t, db.Find(&allMedia).Error) {
		return
	}

	assert.Equal(t, 9, len(allMedia))

	var allMediaURL []*models.MediaURL
	if !assert.NoError(t, db.Find(&allMediaURL).Error) {
		return
	}

	assert.Equal(t, 18, len(allMediaURL))

	// Verify that faces was recognized
	assert.Eventually(t, func() bool {
		var allFaceGroups []*models.FaceGroup
		if !assert.NoError(t, db.Find(&allFaceGroups).Error) {
			return false
		}

		return len(allFaceGroups) == 3
	}, time.Second*5, time.Millisecond*500)

	assert.Eventually(t, func() bool {
		var allImageFaces []*models.ImageFace
		if !assert.NoError(t, db.Find(&allImageFaces).Error) {
			return false
		}

		return len(allImageFaces) == 6
	}, time.Second*5, time.Millisecond*500)

}
