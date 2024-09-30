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

const faceGroupIDisQuestion = "face_group_id = ?"
const mediaAlbumIDinQuestion = "media.album_id IN (?)"
const imageFacesIDinQuestion = "image_faces.id IN (?)"

var ErrFaceDetectorNotInitialized = errors.New("face detector not initialized")

type imageFaceResolver struct {
	*Resolver
}

type faceGroupResolver struct {
	*Resolver
}

func (r *Resolver) ImageFace() api.ImageFaceResolver {
	return imageFaceResolver{r}
}

func (r *Resolver) FaceGroup() api.FaceGroupResolver {
	return faceGroupResolver{r}
}

func (r imageFaceResolver) FaceGroup(ctx context.Context, obj *models.ImageFace) (*models.FaceGroup, error) {
	if obj.FaceGroup != nil {
		return obj.FaceGroup, nil
	}

	if face_detection.GlobalFaceDetector == nil {
		return nil, ErrFaceDetectorNotInitialized
	}

	var faceGroup models.FaceGroup
	if err := r.DB(ctx).Model(&obj).Association("FaceGroup").Find(&faceGroup); err != nil {
		return nil, err
	}

	obj.FaceGroup = &faceGroup

	return &faceGroup, nil
}

func (r imageFaceResolver) Media(ctx context.Context, obj *models.ImageFace) (*models.Media, error) {
	if err := obj.FillMedia(r.DB(ctx)); err != nil {
		return nil, err
	}

	return &obj.Media, nil
}

func (r faceGroupResolver) ImageFaces(ctx context.Context, obj *models.FaceGroup,
	paginate *models.Pagination) ([]*models.ImageFace, error) {
	db := r.DB(ctx)
	user := auth.UserFromContext(ctx)
	if user == nil {
		return nil, errors.New("unauthorized")
	}

	if face_detection.GlobalFaceDetector == nil {
		return nil, ErrFaceDetectorNotInitialized
	}

	if err := user.FillAlbums(db); err != nil {
		return nil, err
	}

	userAlbumIDs := make([]int, len(user.Albums))
	for i, album := range user.Albums {
		userAlbumIDs[i] = album.ID
	}

	query := db.
		Joins("Media").
		Where(faceGroupIDisQuestion, obj.ID).
		Where("album_id IN (?)", userAlbumIDs)

	query = models.FormatSQL(query, nil, paginate)

	var imageFaces []*models.ImageFace
	if err := query.Find(&imageFaces).Error; err != nil {
		return nil, err
	}

	return imageFaces, nil
}

func (r faceGroupResolver) ImageFaceCount(ctx context.Context, obj *models.FaceGroup) (int, error) {
	db := r.DB(ctx)
	user := auth.UserFromContext(ctx)
	if user == nil {
		return -1, errors.New("unauthorized")
	}

	if face_detection.GlobalFaceDetector == nil {
		return -1, ErrFaceDetectorNotInitialized
	}

	if err := user.FillAlbums(db); err != nil {
		return -1, err
	}

	userAlbumIDs := make([]int, len(user.Albums))
	for i, album := range user.Albums {
		userAlbumIDs[i] = album.ID
	}

	query := db.
		Model(&models.ImageFace{}).
		Joins("Media").
		Where(faceGroupIDisQuestion, obj.ID).
		Where("album_id IN (?)", userAlbumIDs)

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return -1, err
	}

	return int(count), nil
}

func (r *queryResolver) FaceGroup(ctx context.Context, id int) (*models.FaceGroup, error) {
	db := r.DB(ctx)
	user := auth.UserFromContext(ctx)
	if user == nil {
		return nil, errors.New("unauthorized")
	}

	if face_detection.GlobalFaceDetector == nil {
		return nil, ErrFaceDetectorNotInitialized
	}

	if err := user.FillAlbums(db); err != nil {
		return nil, err
	}

	userAlbumIDs := make([]int, len(user.Albums))
	for i, album := range user.Albums {
		userAlbumIDs[i] = album.ID
	}

	faceGroupQuery := db.
		Joins("LEFT JOIN image_faces ON image_faces.face_group_id = face_groups.id").
		Joins("LEFT JOIN media ON image_faces.media_id = media.id").
		Where("face_groups.id = ?", id).
		Where(mediaAlbumIDinQuestion, userAlbumIDs)

	var faceGroup models.FaceGroup
	if err := faceGroupQuery.Find(&faceGroup).Error; err != nil {
		return nil, err
	}

	return &faceGroup, nil
}

func (r *queryResolver) MyFaceGroups(ctx context.Context, paginate *models.Pagination) ([]*models.FaceGroup, error) {
	db := r.DB(ctx)
	user := auth.UserFromContext(ctx)
	if user == nil {
		return nil, errors.New("unauthorized")
	}

	if face_detection.GlobalFaceDetector == nil {
		return nil, ErrFaceDetectorNotInitialized
	}

	if err := user.FillAlbums(db); err != nil {
		return nil, err
	}

	userAlbumIDs := make([]int, len(user.Albums))
	for i, album := range user.Albums {
		userAlbumIDs[i] = album.ID
	}

	faceGroupQuery := db.
		Joins("JOIN image_faces ON image_faces.face_group_id = face_groups.id").
		Where("image_faces.media_id IN (?)",
			db.Select("media.id").Table("media").Where(mediaAlbumIDinQuestion, userAlbumIDs)).
		Group("image_faces.face_group_id").
		Group("face_groups.id").
		Order("CASE WHEN label IS NULL THEN 1 ELSE 0 END").
		Order("COUNT(image_faces.id) DESC")

	faceGroupQuery = models.FormatSQL(faceGroupQuery, nil, paginate)

	var faceGroups []*models.FaceGroup
	if err := faceGroupQuery.Find(&faceGroups).Error; err != nil {
		return nil, err
	}

	return faceGroups, nil
}

func (r *mutationResolver) SetFaceGroupLabel(ctx context.Context, faceGroupID int,
	label *string) (*models.FaceGroup, error) {
	db := r.DB(ctx)
	user := auth.UserFromContext(ctx)
	if user == nil {
		return nil, errors.New("unauthorized")
	}

	if face_detection.GlobalFaceDetector == nil {
		return nil, ErrFaceDetectorNotInitialized
	}

	faceGroup, err := userOwnedFaceGroup(db, user, faceGroupID)
	if err != nil {
		return nil, err
	}

	if err := db.Model(faceGroup).Update("label", label).Error; err != nil {
		return nil, err
	}

	return faceGroup, nil
}

func (r *mutationResolver) CombineFaceGroups(ctx context.Context, destinationFaceGroupID int,
	sourceFaceGroupID int) (*models.FaceGroup, error) {
	db := r.DB(ctx)
	user := auth.UserFromContext(ctx)
	if user == nil {
		return nil, errors.New("unauthorized")
	}

	if face_detection.GlobalFaceDetector == nil {
		return nil, ErrFaceDetectorNotInitialized
	}

	destinationFaceGroup, err := userOwnedFaceGroup(db, user, destinationFaceGroupID)
	if err != nil {
		return nil, err
	}

	sourceFaceGroup, err := userOwnedFaceGroup(db, user, sourceFaceGroupID)
	if err != nil {
		return nil, err
	}

	updateError := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.ImageFace{}).
			Where(faceGroupIDisQuestion, sourceFaceGroup.ID).
			Update("face_group_id", destinationFaceGroup.ID).Error; err != nil {
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

func (r *mutationResolver) MoveImageFaces(ctx context.Context, imageFaceIDs []int,
	destinationFaceGroupID int) (*models.FaceGroup, error) {
	db := r.DB(ctx)
	user := auth.UserFromContext(ctx)
	if user == nil {
		return nil, errors.New("unauthorized")
	}

	if face_detection.GlobalFaceDetector == nil {
		return nil, ErrFaceDetectorNotInitialized
	}

	userOwnedImageFaceIDs := make([]int, 0)
	var destFaceGroup *models.FaceGroup

	transErr := db.Transaction(func(tx *gorm.DB) error {

		var err error
		destFaceGroup, err = userOwnedFaceGroup(tx, user, destinationFaceGroupID)
		if err != nil {
			return err
		}

		userOwnedImageFaces, err := getUserOwnedImageFaces(tx, user, imageFaceIDs)
		if err != nil {
			return err
		}

		for _, imageFace := range userOwnedImageFaces {
			userOwnedImageFaceIDs = append(userOwnedImageFaceIDs, imageFace.ID)
		}

		var sourceFaceGroups []*models.FaceGroup
		if err := tx.
			Joins("LEFT JOIN image_faces ON image_faces.face_group_id = face_groups.id").
			Where(imageFacesIDinQuestion, userOwnedImageFaceIDs).
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
		if err := deleteEmptyFaceGroups(sourceFaceGroups, tx); err != nil {
			return err
		}

		return nil
	})

	if transErr != nil {
		return nil, transErr
	}

	face_detection.GlobalFaceDetector.MergeImageFaces(userOwnedImageFaceIDs, int32(destFaceGroup.ID))

	return destFaceGroup, nil
}

func deleteEmptyFaceGroups(sourceFaceGroups []*models.FaceGroup, tx *gorm.DB) error {
	for _, faceGroup := range sourceFaceGroups {
		var count int64
		if err := tx.Model(&models.ImageFace{}).Where(faceGroupIDisQuestion, faceGroup.ID).Count(&count).Error; err != nil {
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

func (r *mutationResolver) RecognizeUnlabeledFaces(ctx context.Context) ([]*models.ImageFace, error) {
	db := r.DB(ctx)
	user := auth.UserFromContext(ctx)
	if user == nil {
		return nil, errors.New("unauthorized")
	}

	if face_detection.GlobalFaceDetector == nil {
		return nil, ErrFaceDetectorNotInitialized
	}

	var updatedImageFaces []*models.ImageFace

	transactionError := db.Transaction(func(tx *gorm.DB) error {
		var err error
		updatedImageFaces, err = face_detection.GlobalFaceDetector.RecognizeUnlabeledFaces(tx, user)

		return err
	})

	if transactionError != nil {
		return nil, transactionError
	}

	return updatedImageFaces, nil
}

func (r *mutationResolver) DetachImageFaces(ctx context.Context, imageFaceIDs []int) (*models.FaceGroup, error) {
	db := r.DB(ctx)
	user := auth.UserFromContext(ctx)
	if user == nil {
		return nil, errors.New("unauthorized")
	}

	if face_detection.GlobalFaceDetector == nil {
		return nil, ErrFaceDetectorNotInitialized
	}

	userOwnedImageFaceIDs := make([]int, 0)
	newFaceGroup := models.FaceGroup{}

	transactionError := db.Transaction(func(tx *gorm.DB) error {

		userOwnedImageFaces, err := getUserOwnedImageFaces(tx, user, imageFaceIDs)
		if err != nil {
			return err
		}

		for _, imageFace := range userOwnedImageFaces {
			userOwnedImageFaceIDs = append(userOwnedImageFaceIDs, imageFace.ID)
		}

		if err := tx.Save(&newFaceGroup).Error; err != nil {
			return err
		}

		if err := tx.
			Model(&models.ImageFace{}).
			Where("id IN (?)", userOwnedImageFaceIDs).
			Update("face_group_id", newFaceGroup.ID).Error; err != nil {
			return err
		}

		return nil
	})

	if transactionError != nil {
		return nil, transactionError
	}

	face_detection.GlobalFaceDetector.MergeImageFaces(userOwnedImageFaceIDs, int32(newFaceGroup.ID))

	return &newFaceGroup, nil
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
		Joins("JOIN media ON media.id = image_faces.media_id").
		Where(mediaAlbumIDinQuestion, userAlbumIDs)

	faceGroupQuery := db.
		Model(&models.FaceGroup{}).
		Joins("JOIN image_faces ON face_groups.id = image_faces.face_group_id").
		Where("face_groups.id = ?", faceGroupID).
		Where(imageFacesIDinQuestion, imageFaceQuery)

	var faceGroup models.FaceGroup
	if err := faceGroupQuery.Find(&faceGroup).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.Wrap(err, "face group does not exist or is not owned by the user")
		}
		return nil, err
	}

	return &faceGroup, nil
}

func getUserOwnedImageFaces(tx *gorm.DB, user *models.User, imageFaceIDs []int) ([]*models.ImageFace, error) {
	if err := user.FillAlbums(tx); err != nil {
		return nil, err
	}

	userAlbumIDs := make([]int, len(user.Albums))
	for i, album := range user.Albums {
		userAlbumIDs[i] = album.ID
	}

	var userOwnedImageFaces []*models.ImageFace
	if err := tx.
		Joins("JOIN media ON media.id = image_faces.media_id").
		Where(mediaAlbumIDinQuestion, userAlbumIDs).
		Where(imageFacesIDinQuestion, imageFaceIDs).
		Find(&userOwnedImageFaces).Error; err != nil {
		return nil, err
	}

	return userOwnedImageFaces, nil
}
