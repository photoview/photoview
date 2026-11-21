package resolvers

import (
	"context"
	"errors"
	"fmt"

	"github.com/photoview/photoview/api/graphql/models"
	"gorm.io/gorm"
)

const faceGroupIDIsQuestion = "face_group_id = ?"
const faceGroupIDsInQuestion = "face_group_id IN (?)"
const imageFacesIDInQuestion = "image_faces.id IN (?)"

// userOwnsImageFace matches image_faces rows whose media is in an album the
// user owns. It is a subquery, so the statement does not bind every album ID.
const userOwnsImageFace = "EXISTS (SELECT 1 FROM media AS owned_media " +
	"JOIN user_albums ON user_albums.album_id = owned_media.album_id " +
	"WHERE owned_media.id = image_faces.media_id AND user_albums.user_id = ?)"

var ErrFaceDetectorNotInitialized = errors.New("face detector not initialized")

func userOwnedFaceGroup(ctx context.Context, db *gorm.DB, user *models.User, faceGroupID int) (*models.FaceGroup, error) {
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

	// Membership is checked in the query. User.Albums is left alone: a websocket
	// connection reuses the user value, and FillAlbums keeps a non-empty slice
	// after access is revoked.
	db = db.WithContext(ctx)
	faceGroupQuery := db.
		Model(&models.FaceGroup{}).
		Joins("JOIN image_faces ON face_groups.id = image_faces.face_group_id").
		Where("face_groups.id = ?", faceGroupID).
		Where(userOwnsImageFace, user.ID)

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

// getUserOwnedImageFaces returns the requested faces the caller may use.
// Administrators receive every requested face, the same exception as move,
// detach, and recognition results. Regular users only receive faces in albums
// they own.
func getUserOwnedImageFaces(tx *gorm.DB, user *models.User, imageFaceIDs []int) ([]*models.ImageFace, error) {
	return lookupImageFaces(tx, user, imageFaceIDs, !user.Admin)
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

func forEachIDBatch(ids []int, fn func(batch []int) error) error {
	for start := 0; start < len(ids); start += imageFaceLookupBatchSize {
		end := min(start+imageFaceLookupBatchSize, len(ids))
		if err := fn(ids[start:end]); err != nil {
			return err
		}
	}
	return nil
}

func uniqueFaceGroups(groups []*models.FaceGroup) []*models.FaceGroup {
	if len(groups) < 2 {
		return groups
	}
	seen := make(map[int]struct{}, len(groups))
	unique := make([]*models.FaceGroup, 0, len(groups))
	for _, group := range groups {
		if group == nil || group.ID == 0 {
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

type mediaCountRow struct {
	MediaID int   `gorm:"column:media_id"`
	Count   int64 `gorm:"column:count"`
}

type faceMediaRow struct {
	ID      int `gorm:"column:id"`
	MediaID int `gorm:"column:media_id"`
}

func hasDuplicateMediaInFaceGroupsUnion(db *gorm.DB, faceGroupIDs ...int) (bool, error) {
	faceGroupIDs = dedupeFaceIDs(faceGroupIDs)
	if len(faceGroupIDs) == 0 {
		return false, nil
	}

	// Count faces per media across batches. One statement that binds every group
	// ID can exceed a database parameter limit when many groups are merged.
	counts := make(map[int]int64)
	err := forEachIDBatch(faceGroupIDs, func(batch []int) error {
		var rows []mediaCountRow
		if err := db.Table("image_faces").
			Select("media_id, COUNT(*) AS count").
			Where("face_group_id IN ?", batch).
			Group("media_id").
			Scan(&rows).Error; err != nil {
			return err
		}
		for _, row := range rows {
			counts[row.MediaID] += row.Count
			if counts[row.MediaID] > 1 {
				return errDuplicateMedia
			}
		}
		return nil
	})
	if errors.Is(err, errDuplicateMedia) {
		return true, nil
	}
	return false, err
}

func doesFaceGroupContainImages(db *gorm.DB, faceGroupID int, imageFaceIDs []int) (bool, error) {
	imageFaceIDs = dedupeFaceIDs(imageFaceIDs)
	moving := make(map[int]struct{}, len(imageFaceIDs))
	for _, id := range imageFaceIDs {
		moving[id] = struct{}{}
	}

	// Faces that will be in the destination: those already there, except faces
	// that are part of this move, plus the faces being moved. The move list is
	// loaded in batches and each ID is bound once. The previous statement bound
	// the same list twice, which can pass a database parameter limit.
	counts := make(map[int]int64)
	var existing []faceMediaRow
	if err := db.Table("image_faces").
		Select("id, media_id").
		Where("face_group_id = ?", faceGroupID).
		Find(&existing).Error; err != nil {
		return false, err
	}
	for _, face := range existing {
		if _, ok := moving[face.ID]; ok {
			continue
		}
		counts[face.MediaID]++
		if counts[face.MediaID] > 1 {
			return true, nil
		}
	}

	err := forEachIDBatch(imageFaceIDs, func(batch []int) error {
		var moved []faceMediaRow
		if err := db.Table("image_faces").
			Select("media_id").
			Where("id IN ?", batch).
			Find(&moved).Error; err != nil {
			return err
		}
		for _, face := range moved {
			counts[face.MediaID]++
			if counts[face.MediaID] > 1 {
				return errDuplicateMedia
			}
		}
		return nil
	})
	if errors.Is(err, errDuplicateMedia) {
		return true, nil
	}
	if err != nil {
		return false, err
	}
	return false, nil
}

var errDuplicateMedia = errors.New("duplicate media in face group")

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
