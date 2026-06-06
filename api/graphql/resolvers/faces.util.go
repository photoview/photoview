package resolvers

import (
	"errors"
	"fmt"

	"github.com/photoview/photoview/api/graphql/models"
	"gorm.io/gorm"
)

const faceGroupIDIsQuestion = "face_group_id = ?"
const faceGroupIDsInQuestion = "face_group_id IN (?)"
const mediaAlbumIDInQuestion = "media.album_id IN (?)"
const imageFacesIDInQuestion = "image_faces.id IN (?)"

var ErrFaceDetectorNotInitialized = errors.New("face detector not initialized")

func userOwnedFaceGroup(db *gorm.DB, user *models.User, faceGroupID int) (*models.FaceGroup, error) {
	if user.Admin {
		var faceGroup models.FaceGroup
		if err := db.Where("id = ?", faceGroupID).Find(&faceGroup).Error; err != nil {
			return nil, err
		}

		return &faceGroup, nil
	}

	if err := user.FillAlbums(db); err != nil {
		return nil, err
	}

	userAlbumIDs := make([]int, len(user.Albums))
	for i, album := range user.Albums {
		userAlbumIDs[i] = album.ID
	}

	// Verify that user owns at least one of the images in the face group
	imageFaceQuery := db.
		Select("image_faces.id").
		Table("image_faces").
		Joins("JOIN media ON media.id = image_faces.media_id").
		Where(mediaAlbumIDInQuestion, userAlbumIDs)

	faceGroupQuery := db.
		Model(&models.FaceGroup{}).
		Joins("JOIN image_faces ON face_groups.id = image_faces.face_group_id").
		Where("face_groups.id = ?", faceGroupID).
		Where(imageFacesIDInQuestion, imageFaceQuery)

	var faceGroup models.FaceGroup
	if err := faceGroupQuery.Find(&faceGroup).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("face group does not exist or is not owned by the user: %w", err)
		}
		return nil, err
	}

	return &faceGroup, nil
}

func getUserOwnedImageFaces(tx *gorm.DB, user *models.User, imageFaceIDs []int) ([]*models.ImageFace, error) {
	var userOwnedImageFaces []*models.ImageFace

	if len(imageFaceIDs) == 0 {
		return userOwnedImageFaces, nil
	}

	query := tx.Model(&models.ImageFace{})

	if !user.Admin {
		if err := user.FillAlbums(tx); err != nil {
			return nil, err
		}

		userAlbumIDs := make([]int, len(user.Albums))
		for i, album := range user.Albums {
			userAlbumIDs[i] = album.ID
		}

		if len(userAlbumIDs) == 0 {
			return userOwnedImageFaces, nil
		}

		query = query.
			Joins("JOIN media ON media.id = image_faces.media_id").
			Where(mediaAlbumIDInQuestion, userAlbumIDs)
	}

	if err := query.
		Where(imageFacesIDInQuestion, imageFaceIDs).
		Find(&userOwnedImageFaces).Error; err != nil {
		return nil, err
	}

	return userOwnedImageFaces, nil
}

func faceGroupsWouldContainDuplicateMedia(db *gorm.DB, destinationFaceGroupID int, sourceFaceGroupIDs []int) (bool, error) {
	faceGroupIDs := append([]int{destinationFaceGroupID}, sourceFaceGroupIDs...)

	duplicateMediaQuery := db.
		Table("image_faces").
		Select("media_id").
		Where("face_group_id IN ?", faceGroupIDs).
		Group("media_id").
		Having("COUNT(*) > 1")

	var count int64
	if err := db.
		Table("(?) AS duplicate_media", duplicateMediaQuery).
		Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

func movingImageFacesWouldCreateDuplicateMedia(db *gorm.DB, destinationFaceGroupID int, imageFaceIDs []int) (bool, error) {
	duplicateMediaQuery := db.
		Table("image_faces AS candidate").
		Select("candidate.media_id").
		Where(
			"(candidate.face_group_id = ? AND candidate.id NOT IN ?) OR candidate.id IN ?",
			destinationFaceGroupID,
			imageFaceIDs,
			imageFaceIDs,
		).
		Group("candidate.media_id").
		Having("COUNT(*) > 1")

	var count int64
	if err := db.
		Table("(?) AS duplicate_media", duplicateMediaQuery).
		Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

func deleteEmptyFaceGroups(sourceFaceGroups []*models.FaceGroup, tx *gorm.DB) error {
	for _, faceGroup := range sourceFaceGroups {
		var count int64
		if err := tx.Model(&models.ImageFace{}).Where(faceGroupIDIsQuestion, faceGroup.ID).Count(&count).Error; err != nil {
			return err
		}

		if count == 0 {
			if err := tx.Delete(&faceGroup).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func deleteFaceGroups(sourceFaceGroups []*models.FaceGroup, tx *gorm.DB) error {
	for _, faceGroup := range sourceFaceGroups {
		if err := tx.Delete(&faceGroup).Error; err != nil {
			return fmt.Errorf("Delete FaceGroup(%d) error: %w", faceGroup.ID, err)
		}
	}
	return nil
}
