package resolvers

import (
	"context"

	api "github.com/photoview/photoview/api/graphql"
	"github.com/photoview/photoview/api/graphql/auth"
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner/face_detection"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type imageFaceResolver struct {
	*Resolver
}

func (r *Resolver) ImageFace() api.ImageFaceResolver {
	return imageFaceResolver{r}
}

func (r imageFaceResolver) FaceGroup(ctx context.Context, obj *models.ImageFace) (*models.FaceGroup, error) {
	if obj.FaceGroup != nil {
		return obj.FaceGroup, nil
	}

	var faceGroup models.FaceGroup
	if err := r.Database.Model(&obj).Association("FaceGroup").Find(&faceGroup); err != nil {
		return nil, err
	}

	obj.FaceGroup = &faceGroup

	return &faceGroup, nil
}

func (r *queryResolver) MyFaceGroups(ctx context.Context, paginate *models.Pagination) ([]*models.FaceGroup, error) {
	user := auth.UserFromContext(ctx)
	if user == nil {
		return nil, errors.New("unauthorized")
	}

	if err := user.FillAlbums(r.Database); err != nil {
		return nil, err
	}

	userAlbumIDs := make([]int, len(user.Albums))
	for i, album := range user.Albums {
		userAlbumIDs[i] = album.ID
	}

	imageFaceQuery := r.Database.
		Joins("Media").
		Where("media.album_id IN (?)", userAlbumIDs)

	var imageFaces []*models.ImageFace
	if err := imageFaceQuery.Find(&imageFaces).Error; err != nil {
		return nil, err
	}

	faceGroupMap := make(map[int][]models.ImageFace)
	for _, face := range imageFaces {
		_, found := faceGroupMap[face.FaceGroupID]

		if found {
			faceGroupMap[face.FaceGroupID] = append(faceGroupMap[face.FaceGroupID], *face)
		} else {
			faceGroupMap[face.FaceGroupID] = make([]models.ImageFace, 1)
			faceGroupMap[face.FaceGroupID][0] = *face
		}
	}

	faceGroupIDs := make([]int, len(faceGroupMap))
	i := 0
	for groupID := range faceGroupMap {
		faceGroupIDs[i] = groupID
		i++
	}

	faceGroupQuery := r.Database.
		Joins("LEFT JOIN image_faces ON image_faces.id = face_groups.id").
		Where("face_groups.id IN (?)", faceGroupIDs).
		Order("CASE WHEN label IS NULL THEN 1 ELSE 0 END")

	var faceGroups []*models.FaceGroup
	if err := faceGroupQuery.Find(&faceGroups).Error; err != nil {
		return nil, err
	}

	for _, faceGroup := range faceGroups {
		faceGroup.ImageFaces = faceGroupMap[faceGroup.ID]
	}

	return faceGroups, nil
}

func (r *mutationResolver) SetFaceGroupLabel(ctx context.Context, faceGroupID int, label *string) (*models.FaceGroup, error) {
	user := auth.UserFromContext(ctx)
	if user == nil {
		return nil, errors.New("unauthorized")
	}

	faceGroup, err := userOwnedFaceGroup(r.Database, user, faceGroupID)
	if err != nil {
		return nil, err
	}

	if err := r.Database.Model(faceGroup).Update("label", label).Error; err != nil {
		return nil, err
	}

	return faceGroup, nil
}

func (r *mutationResolver) CombineFaceGroups(ctx context.Context, destinationFaceGroupID int, sourceFaceGroupID int) (*models.FaceGroup, error) {
	user := auth.UserFromContext(ctx)
	if user == nil {
		return nil, errors.New("unauthorized")
	}

	destinationFaceGroup, err := userOwnedFaceGroup(r.Database, user, destinationFaceGroupID)
	if err != nil {
		return nil, err
	}

	sourceFaceGroup, err := userOwnedFaceGroup(r.Database, user, sourceFaceGroupID)
	if err != nil {
		return nil, err
	}

	updateError := r.Database.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.ImageFace{}).Where("face_group_id = ?", sourceFaceGroup.ID).Update("face_group_id", destinationFaceGroup.ID).Error; err != nil {
			return err
		}

		if err := tx.Delete(&sourceFaceGroup).Error; err != nil {
			return err
		}

		return nil
	})

	if updateError != nil {
		return nil, updateError
	}

	face_detection.GlobalFaceDetector.MergeCategories(int32(sourceFaceGroupID), int32(destinationFaceGroupID))

	return destinationFaceGroup, nil
}

func (r *mutationResolver) MoveImageFaces(ctx context.Context, imageFaceIDs []int, destinationFaceGroupID int) (*models.FaceGroup, error) {
	user := auth.UserFromContext(ctx)
	if user == nil {
		return nil, errors.New("unauthorized")
	}

	if err := user.FillAlbums(r.Database); err != nil {
		return nil, err
	}

	userAlbumIDs := make([]int, len(user.Albums))
	for i, album := range user.Albums {
		userAlbumIDs[i] = album.ID
	}

	userOwnedImageFaceIDs := make([]int, 0)
	var destFaceGroup *models.FaceGroup

	transErr := r.Database.Transaction(func(tx *gorm.DB) error {

		var err error
		destFaceGroup, err = userOwnedFaceGroup(tx, user, destinationFaceGroupID)
		if err != nil {
			return err
		}

		var userOwnedImageFaces []*models.ImageFace
		if err := tx.
			Joins("JOIN media ON media.id = image_faces.media_id").
			Where("media.album_id IN (?)", userAlbumIDs).
			Where("image_faces.id IN (?)", imageFaceIDs).
			Find(&userOwnedImageFaces).Error; err != nil {
			return err
		}

		for _, imageFace := range userOwnedImageFaces {
			userOwnedImageFaceIDs = append(userOwnedImageFaceIDs, imageFace.ID)
		}

		var sourceFaceGroups []*models.FaceGroup
		if err := tx.
			Joins("LEFT JOIN image_faces ON image_faces.face_group_id = face_groups.id").
			Where("image_faces.id IN (?)", userOwnedImageFaceIDs).
			Find(&sourceFaceGroups).Error; err != nil {
			return err
		}

		if err := tx.
			Model(&models.ImageFace{}).
			Where("id IN (?)", userOwnedImageFaceIDs).
			Update("face_group_id", destFaceGroup.ID).Error; err != nil {
			return err
		}

		// delete face groups if they have become empty
		for _, faceGroup := range sourceFaceGroups {
			var count int64
			if err := tx.Model(&models.ImageFace{}).Where("face_group_id = ?", faceGroup.ID).Count(&count).Error; err != nil {
				return err
			}

			if count == 0 {
				if err := tx.Delete(&faceGroup).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})

	if transErr != nil {
		return nil, transErr
	}

	face_detection.GlobalFaceDetector.MergeImageFaces(userOwnedImageFaceIDs, int32(destFaceGroup.ID))

	return destFaceGroup, nil
}

func (r *mutationResolver) RecognizeUnlabeledFaces(ctx context.Context) ([]*models.ImageFace, error) {
	user := auth.UserFromContext(ctx)
	if user == nil {
		return nil, errors.New("unauthorized")
	}

	var updatedImageFaces []*models.ImageFace

	transactionError := r.Database.Transaction(func(tx *gorm.DB) error {
		var err error
		updatedImageFaces, err = face_detection.GlobalFaceDetector.RecognizeUnlabeledFaces(tx, user)

		return err
	})

	if transactionError != nil {
		return nil, transactionError
	}

	return updatedImageFaces, nil
}

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

	// Verify that user owns at leat one of the images in the face group
	imageFaceQuery := db.
		Select("image_faces.id").
		Table("image_faces").
		Joins("LEFT JOIN media ON media.id = image_faces.media_id").
		Where("media.album_id IN (?)", userAlbumIDs)

	faceGroupQuery := db.
		Model(&models.FaceGroup{}).
		Joins("JOIN image_faces ON face_groups.id = image_faces.face_group_id").
		Where("face_groups.id = ?", faceGroupID).
		Where("image_faces.id IN (?)", imageFaceQuery)

	var faceGroup models.FaceGroup
	if err := faceGroupQuery.Find(&faceGroup).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.Wrap(err, "face group does not exist or is not owned by the user")
		}
		return nil, err
	}

	return &faceGroup, nil
}
