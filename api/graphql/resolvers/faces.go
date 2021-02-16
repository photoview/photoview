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

	var imageFaces []*models.ImageFace
	if err := r.Database.Joins("Media").Where("media.album_id IN (?)", userAlbumIDs).Find(&imageFaces).Error; err != nil {
		return nil, err
	}

	faceGroupMap := make(map[int][]models.ImageFace)
	for _, face := range imageFaces {
		group, found := faceGroupMap[face.FaceGroupID]

		if found {
			group = append(group, *face)
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

	var faceGroups []*models.FaceGroup
	if err := r.Database.Where("id IN (?)", faceGroupIDs).Find(&faceGroups).Error; err != nil {
		return nil, err
	}

	for _, faceGroup := range faceGroups {
		faceGroup.ImageFaces = faceGroupMap[faceGroup.ID]
	}

	return faceGroups, nil
}
