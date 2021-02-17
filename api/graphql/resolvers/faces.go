package resolvers

import (
	"context"
	"errors"

	"github.com/photoview/photoview/api/graphql/auth"
	"github.com/photoview/photoview/api/graphql/models"
)

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

	if err := user.FillAlbums(r.Database); err != nil {
		return nil, err
	}

	userAlbumIDs := make([]int, len(user.Albums))
	for i, album := range user.Albums {
		userAlbumIDs[i] = album.ID
	}

	// Verify that user owns at leat one of the images in the face group
	imageFaceQuery := r.Database.
		Select("image_faces.id").
		Table("image_faces").
		Joins("LEFT JOIN media ON media.id = image_faces.media_id").
		Where("media.album_id IN (?)", userAlbumIDs)

	faceGroupQuery := r.Database.
		Model(&models.FaceGroup{}).
		Joins("JOIN image_faces ON face_groups.id = image_faces.face_group_id").
		Where("face_groups.id = ?", faceGroupID).
		Where("image_faces.id IN (?)", imageFaceQuery)

	var faceGroup models.FaceGroup
	if err := faceGroupQuery.Find(&faceGroup).Error; err != nil {
		return nil, err
	}

	if err := r.Database.Model(&faceGroup).Update("label", label).Error; err != nil {
		return nil, err
	}

	return &faceGroup, nil
}

func (r *mutationResolver) CombineFaceGroups(ctx context.Context, destinationFaceGroupID int, sourceFaceGroupID int) (*models.FaceGroup, error) {
	panic("not implemented")
}

func (r *mutationResolver) MoveImageFace(ctx context.Context, imageFaceID int, newFaceGroupID int) (*models.ImageFace, error) {
	panic("not implemented")
}

func (r *mutationResolver) RecognizeUnlabeledFaces(ctx context.Context) ([]*models.ImageFace, error) {
	panic("not implemented")
}
