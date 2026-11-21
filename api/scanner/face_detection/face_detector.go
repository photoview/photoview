package face_detection

import (
	"errors"

	"gorm.io/gorm"

	"github.com/photoview/photoview/api/graphql/models"
)

type FaceDetector interface {
	ReloadFacesFromDatabase(db *gorm.DB) error
	DetectFaces(db *gorm.DB, media *models.Media) error
	MergeCategories(sourceID int32, destID int32)
	MergeImageFaces(imageFaceIDs []int, destFaceGroupID int32)
	RecognizeUnlabeledFaces(tx *gorm.DB, user *models.User) ([]*models.ImageFace, error)
}

var GlobalFaceDetector FaceDetector = nil

// UnlabeledFaceGroups returns unlabeled face groups the caller may recognize.
// Administrators see every unlabeled group, including groups in albums they do
// not belong to. Other users only see groups that contain a face in an album
// they own. Membership is a subquery, so the statement does not bind one
// parameter per album.
func UnlabeledFaceGroups(tx *gorm.DB, user *models.User) ([]*models.FaceGroup, error) {
	if user == nil {
		return nil, errors.New("unauthorized")
	}

	query := tx.
		Joins("JOIN image_faces ON image_faces.face_group_id = face_groups.id").
		Joins("JOIN media ON image_faces.media_id = media.id").
		Where("face_groups.label IS NULL")
	if !user.Admin {
		query = query.Where(
			"EXISTS (SELECT 1 FROM user_albums WHERE user_albums.album_id = media.album_id AND user_albums.user_id = ?)",
			user.ID,
		)
	}

	var groups []*models.FaceGroup
	if err := query.Find(&groups).Error; err != nil {
		return nil, err
	}
	return uniqueFaceGroups(groups), nil
}

func uniqueFaceGroups(groups []*models.FaceGroup) []*models.FaceGroup {
	if len(groups) < 2 {
		return groups
	}
	seen := make(map[int]struct{}, len(groups))
	unique := make([]*models.FaceGroup, 0, len(groups))
	for _, group := range groups {
		if group == nil {
			continue
		}
		if _, ok := seen[group.ID]; ok {
			continue
		}
		seen[group.ID] = struct{}{}
		unique = append(unique, group)
	}
	return unique
}
