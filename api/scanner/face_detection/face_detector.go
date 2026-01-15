package face_detection

import (
	"gorm.io/gorm"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/spf13/afero"
)

type FaceDetector interface {
	ReloadFacesFromDatabase(db *gorm.DB) error
	DetectFaces(db *gorm.DB, fs afero.Fs, media *models.Media) error
	MergeCategories(sourceID int32, destID int32)
	MergeImageFaces(imageFaceIDs []int, destFaceGroupID int32)
	RecognizeUnlabeledFaces(tx *gorm.DB, user *models.User) ([]*models.ImageFace, error)
}

var GlobalFaceDetector FaceDetector = nil
