package resolvers

import (
	"context"
	"testing"

	"github.com/photoview/photoview/api/graphql/auth"
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner/face_detection"
	"github.com/photoview/photoview/api/test_utils"
)

func TestCombineFaceGroups(t *testing.T) {
	test_utils.FilesystemTest(t)
	db := test_utils.DatabaseTest(t)
	face_detection.InitializeFaceDetector(db)
	pass := "1234"
	user, err := models.RegisterUser(db, "test_user", &pass, true)
	if err != nil {
		t.Fatal("register user error:", err)
	}
	db.AutoMigrate(&models.ImageFace{}, &models.FaceGroup{}, &models.Media{}, &models.Album{})
	tests := []struct {
		name string
		dest int
		src  []int
	}{
		{
			name: "merge multiple combinations with duplicates",
			dest: 1,
			src:  []int{2, 3},
		},
		{
			name: "merge two combinations with duplicates",
			dest: 1,
			src:  []int{2},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			db.Exec("DELETE FROM image_faces")
			db.Exec("DELETE FROM face_groups")
			db.Exec("DELETE FROM media")
			db.Exec("DELETE FROM albums")

			testAlbum := models.Album{Title: "Test Album"}
			if err := db.Create(&testAlbum).Error; err != nil {
				t.Fatal(err)
			}

			testMedia := []models.Media{
				{Model: models.Model{ID: 1}, Path: "test1", AlbumID: testAlbum.ID},
				{Model: models.Model{ID: 2}, Path: "test2", AlbumID: testAlbum.ID},
				{Model: models.Model{ID: 3}, Path: "test3", AlbumID: testAlbum.ID},
				{Model: models.Model{ID: 4}, Path: "test4", AlbumID: testAlbum.ID},
			}
			if err := db.Create(&testMedia).Error; err != nil {
				t.Fatal(err)
			}

			testFaceGroup := []models.FaceGroup{
				{Model: models.Model{ID: 1}},
				{Model: models.Model{ID: 2}},
				{Model: models.Model{ID: 3}},
				{Model: models.Model{ID: 4}},
			}
			if err := db.Create(&testFaceGroup).Error; err != nil {
				t.Fatal(err)
			}

			testDataList := []models.ImageFace{
				{FaceGroupID: 1, MediaID: 1},
				{FaceGroupID: 1, MediaID: 2},
				{FaceGroupID: 1, MediaID: 3},
				{FaceGroupID: 2, MediaID: 3},
				{FaceGroupID: 2, MediaID: 4},
				{FaceGroupID: 3, MediaID: 4},
				{FaceGroupID: 3, MediaID: 1},
			}
			if err := db.Create(&testDataList).Error; err != nil {
				t.Fatal(err)
			}

			r := &mutationResolver{
				Resolver: &Resolver{
					database: db,
				},
			}
			ctx := auth.AddUserToContext(context.Background(), user)

			combineFace, err := r.CombineFaceGroups(ctx, tt.dest, tt.src)
			if err != nil {
				t.Fatal("test CombineFaceGroups err:", err)
			}

			m := make(map[int]struct{})
			for _, imageface := range combineFace.ImageFaces {
				if _, ok := m[imageface.MediaID]; ok {
					t.Fatal("filtering failed at", imageface.MediaID)
				}
				m[imageface.MediaID] = struct{}{}
			}
		})
	}
}
