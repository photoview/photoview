package resolvers_test

import (
	// "context"
	"testing"

	api "github.com/photoview/photoview/api/graphql"
	// "github.com/photoview/photoview/api/graphql/auth"
	// "./../generated"
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/graphql/resolvers"
	// "github.com/pkg/errors"
	// "gorm.io/gorm"

	"github.com/photoview/photoview/api/test_utils"
	"github.com/stretchr/testify/assert"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/client"
)

// func NewResolver() api.Config {
// 	r := Resolver{}
//
// 	r.Database = test_utils.DatabaseTest(t)
//
// 	return api.Config{
// 		Resolvers: &r,
// 	}
// }

func TestAlbumCover(t *testing.T) {
	db := test_utils.DatabaseTest(t)

	rootAlbum := models.Album{
		Title: "root",
		Path:  "/photos",
	}

	if !assert.NoError(t, db.Save(&rootAlbum).Error) {
		return
	}

	TestCoverID := 6

	children := []models.Album{
		{
			Title:         "child1",
			Path:          "/photos/child1",
			ParentAlbumID: &rootAlbum.ID,
		},
		{
			Title:         "child2",
			Path:          "/photos/child2",
			ParentAlbumID: &rootAlbum.ID,
			CoverID:       &TestCoverID,
		},
	}

	if !assert.NoError(t, db.Save(&children).Error) {
		return
	}

	photos := []models.Media{
		{
			Title:   "pic1",
			Path:    "/photos/pic1",
			AlbumID: 1,
		},
		{
			Title:   "pic2",
			Path:    "/photos/pic2",
			AlbumID: 1,
		},
		{
			Title:   "pic3",
			Path:    "/photos/pic3",
			AlbumID: 2,
		},
		{
			Title:   "pic4",
			Path:    "/photos/pic4",
			AlbumID: 2,
		},
		{
			Title:   "pic5",
			Path:    "/photos/pic5",
			AlbumID: 3,
		},
		{
			Title:   "pic6",
			Path:    "/photos/pic6",
			AlbumID: 3,
		},
	}

	if !assert.NoError(t, db.Save(&photos).Error) {
		return
	}

	// verifyResult := func(t *testing.T, expected_albums []*models.Album, result []*models.Album) {
	// 	assert.Equal(t, len(expected_albums), len(result))
	//
	// 	for _, expected := range expected_albums {
	// 		found_expected := false
	// 		for _, item := range result {
	// 			if item.Title == expected.Title && item.Path == expected.Path {
	// 				found_expected = true
	// 				break
	// 			}
	// 		}
	// 		if !found_expected {
	// 			assert.Failf(t, "albums did not match", "expected to find item: %v", expected)
	// 		}
	// 	}
	// }
	// //

	c := client.New(handler.NewDefaultServer(api.NewExecutableSchema(api.Config{Resolvers: &resolvers.Resolver{
		Database: db,
	}})))

	// t.Run("Album get cover photos", func(t *testing.T) {
	// 	root_children, err := rootAlbum.GetChildren(db, nil)
	// 	if !assert.NoError(t, err) {
	// 		return
	// 	}
	//
	// 	expected_children := []*models.Album{
	// 		{
	// 			Title: "root",
	// 			Path:  "/photos",
	// 		},
	// 		{
	// 			Title: "child1",
	// 			Path:  "/photos/child1",
	// 		},
	// 		{
	// 			Title: "child2",
	// 			Path:  "/photos/child2",
	// 		},
	// 		{
	// 			Title: "subchild",
	// 			Path:  "/photos/child1/subchild",
	// 		},
	// 	}
	//
	// 	verifyResult(t, expected_children, root_children)
	// })

	// t.Run("Album get parents", func(t *testing.T) {
	// 	parents, err := sub_child.GetParents(db, nil)
	// 	if !assert.NoError(t, err) {
	// 		return
	// 	}
	//
	// 	expected_parents := []*models.Album{
	// 		{
	// 			Title: "root",
	// 			Path:  "/photos",
	// 		},
	// 		{
	// 			Title: "child1",
	// 			Path:  "/photos/child1",
	// 		},
	// 		{
	// 			Title: "subchild",
	// 			Path:  "/photos/child1/subchild",
	// 		},
	// 	}
	//
	// 	verifyResult(t, expected_parents, parents)
	// })

}

//
//
// func (r *albumResolver) Thumbnail(ctx context.Context, obj *models.Album) (*models.Media, error) {
//
// 	var media models.Media
//
// 	fmt.Print(obj.CoverID)
//
// 	if obj.CoverID == nil {
// 		if err := r.Database.Raw(`
// 			WITH recursive sub_albums AS (
// 				SELECT * FROM albums AS root WHERE id = ?
// 				UNION ALL
// 				SELECT child.* FROM albums AS child JOIN sub_albums ON child.parent_album_id = sub_albums.id
// 			)
//
// 			SELECT * FROM media WHERE media.album_id IN (
// 				SELECT id FROM sub_albums
// 			) AND media.id IN (
// 				SELECT media_id FROM media_urls WHERE media_urls.media_id = media.id
// 			) LIMIT 1
// 		`, obj.ID).Find(&media).Error; err != nil {
// 			return nil, err
// 		}
// 	} else {
// 		if err := r.Database.Where("id = ?", obj.CoverID).Find(&media).Error; err != nil {
// 			return nil, err
// 		}
// 	}
//
// 	return &media, nil
// }
//
// // Takes album_id, resets album.cover_id to 0 (null)
// func (r *mutationResolver) ResetAlbumCover(ctx context.Context, albumID int) (*models.Album, error) {
// 	user := auth.UserFromContext(ctx)
// 	if user == nil {
// 		return nil, errors.New("unauthorized")
// 	}
//
// 	var album models.Album
// 	if err := r.Database.Find(&album, albumID).Error; err != nil {
// 		return nil, err
// 	}
//
// 	ownsAlbum, err := user.OwnsAlbum(r.Database, &album)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	if !ownsAlbum {
// 		return nil, errors.New("forbidden")
// 	}
//
// 	if err := r.Database.Model(&album).Update("cover_id", nil).Error; err != nil {
// 		return nil, err
// 	}
//
// 	return &album, nil
// }
//
// // Takes media.id, finds parent album, sets album.cover_id to media.id (must be a more efficient way of doing this, but it works)
// func (r *mutationResolver) SetAlbumCover(ctx context.Context, coverID int) (*models.Album, error) {
// 	user := auth.UserFromContext(ctx)
// 	if user == nil {
// 		return nil, errors.New("unauthorized")
// 	}
//
// 	var media models.Media
//
// 	if err := r.Database.Find(&media, coverID).Error; err != nil {
// 		return nil, err
// 	}
//
// 	var album models.Album
//
// 	if err := r.Database.Find(&album, &media.AlbumID).Error; err != nil {
// 		return nil, err
// 	}
//
// 	ownsAlbum, err := user.OwnsAlbum(r.Database, &album)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	if !ownsAlbum {
// 		return nil, errors.New("forbidden")
// 	}
//
// 	if err := r.Database.Model(&album).Update("cover_id", coverID).Error; err != nil {
// 		return nil, err
// 	}
//
// 	return &album, nil
// }
