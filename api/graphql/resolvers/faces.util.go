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
		if err := db.Where("id = ?", faceGroupID).Take(&faceGroup).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, fmt.Errorf("face group does not exist: %w", err)
			}
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
	if err := faceGroupQuery.Take(&faceGroup).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("face group does not exist or is not owned by the user: %w", err)
		}
		return nil, err
	}

	return &faceGroup, nil
}

// imageFaceLookupBatchSize bounds how many face IDs one lookup binds. Album
// membership is a user_albums subquery, so the statement stays near this size
// however many albums the user owns. SQLite's default variable limit is 32766.
const imageFaceLookupBatchSize = 1000

func getUserOwnedImageFaces(tx *gorm.DB, user *models.User, imageFaceIDs []int) ([]*models.ImageFace, error) {
	// Move and detach keep the existing administrator exception.
	return lookupImageFaces(tx, user, imageFaceIDs, !user.Admin)
}

// imageFacesInUserAlbums filters recognition results to albums the caller owns.
// Administrators are not exempt: ImageFace.media does not repeat the media query's
// membership check.
func imageFacesInUserAlbums(tx *gorm.DB, user *models.User, imageFaceIDs []int) ([]*models.ImageFace, error) {
	return lookupImageFaces(tx, user, imageFaceIDs, true)
}

func lookupImageFaces(tx *gorm.DB, user *models.User, imageFaceIDs []int, restrictToUserAlbums bool) ([]*models.ImageFace, error) {
	if len(imageFaceIDs) == 0 {
		return nil, nil
	}

	// A single IN query returns each row once. Concatenating batch queries would
	// repeat an ID that appears in more than one batch, and later statements bind
	// that list again.
	imageFaceIDs = dedupeFaceIDs(imageFaceIDs)

	var userOwnedImageFaces []*models.ImageFace
	for start := 0; start < len(imageFaceIDs); start += imageFaceLookupBatchSize {
		end := min(start+imageFaceLookupBatchSize, len(imageFaceIDs))

		query := tx.Model(&models.ImageFace{})
		if restrictToUserAlbums {
			query = query.
				Joins("JOIN media ON media.id = image_faces.media_id").
				Where(
					"EXISTS (SELECT 1 FROM user_albums WHERE user_albums.album_id = media.album_id AND user_albums.user_id = ?)",
					user.ID,
				)
		}

		var batch []*models.ImageFace
		if err := query.
			Where(imageFacesIDInQuestion, imageFaceIDs[start:end]).
			Find(&batch).Error; err != nil {
			return nil, err
		}

		userOwnedImageFaces = append(userOwnedImageFaces, batch...)
	}

	return userOwnedImageFaces, nil
}

func dedupeFaceIDs(ids []int) []int {
	if len(ids) < 2 {
		return ids
	}

	seen := make(map[int]struct{}, len(ids))
	unique := make([]int, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		unique = append(unique, id)
	}
	return unique
}

func hasDuplicateMediaInFaceGroupsUnion(db *gorm.DB, faceGroupIDs ...int) (bool, error) {

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

func doesFaceGroupContainImages(db *gorm.DB, faceGroupID int, imageFaceIDs []int) (bool, error) {
	duplicateMediaQuery := db.
		Table("image_faces AS candidate").
		Select("candidate.media_id").
		Where(
			"(candidate.face_group_id = ? AND candidate.id NOT IN ?) OR candidate.id IN ?",
			faceGroupID,
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
