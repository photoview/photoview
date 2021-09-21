package resolvers_test

import (
	"testing"

	api "github.com/photoview/photoview/api/graphql"
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/graphql/resolvers"

	"github.com/99designs/gqlgen/client"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/photoview/photoview/api/test_utils"
	"github.com/stretchr/testify/assert"
)

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
			AlbumID: &rootAlbum.ID,
		},
		{
			Title:   "pic2",
			Path:    "/photos/pic2",
			AlbumID: &rootAlbum.ID,
		},
		{
			Title:   "pic3",
			Path:    "/photos/child1/pic3",
			AlbumID: &children[0].ID,
		},
		{
			Title:   "pic4",
			Path:    "/photos/child1/pic4",
			AlbumID: &children[0].ID,
		},
		{
			Title:   "pic5",
			Path:    "/photos/child2/pic5",
			AlbumID: &children[1].ID,
		},
		{
			Title:   "pic6",
			Path:    "/photos/child2/pic6",
			AlbumID: &children[1].ID,
		},
	}

	if !assert.NoError(t, db.Save(&photos).Error) {
		return
	}

	verifyResult := func(t *testing.T, expected_media []*models.Media, result []*models.Media) {
		assert.Equal(t, len(expected_media), len(result))

		for _, expected := range expected_media {
			found_expected := false
			for _, item := range result {
				if item.Title == expected.Title && item.Path == expected.Path && item.AlbumID == expected.AlbumID {
					found_expected = true
					break
				}
			}
			if !found_expected {
				assert.Failf(t, "media did not match", "expected to find item: %v", expected)
			}
		}
	}
	//

	c := client.New(handler.NewDefaultServer(api.NewExecutableSchema(api.Config{Resolvers: &resolvers.Resolver{
		Database: db,
	}})))

	t.Run("Album get cover photos", func(t *testing.T) {

		var resp []*models.Media

		c.MustPost(`query { Thumbnail( album(id:1)) {
			Title, Path, AlbumID
		}}`, &resp[0])
		c.MustPost(`query { Thumbnail( album(id:1)) {
			Title, Path, AlbumID
		}}`, &resp[1])
		c.MustPost(`query { Thumbnail( album(id:1)) {
			Title, Path, AlbumID
		}}`, &resp[2])

		expected_thumbnails := []*models.Media{
			{
				Title:   "pic1",
				Path:    "/photos/pic1",
				AlbumID: 1,
			},
			{
				Title:   "pic3",
				Path:    "/photos/child1/pic3",
				AlbumID: 2,
			},
			{
				Title:   "pic6",
				Path:    "/photos/child2/pic6",
				AlbumID: 3,
			},
		}

		verifyResult(t, expected_thumbnails, resp)
	})

	t.Run("Album change cover photos", func(t *testing.T) {

		var resp []*models.Media

		c.MustPost(`query { Thumbnail( album(id:1)) {
			Title, Path, AlbumID
		}}`, &resp[0])
		c.MustPost(`query { Thumbnail( album(id:1)) {
			Title, Path, AlbumID
		}}`, &resp[1])
		c.MustPost(`query { Thumbnail( album(id:1)) {
			Title, Path, AlbumID
		}}`, &resp[2])

		expected_thumbnails := []*models.Media{
			{
				Title:   "pic1",
				Path:    "/photos/pic1",
				AlbumID: 1,
			},
			{
				Title:   "pic3",
				Path:    "/photos/child1/pic3",
				AlbumID: 2,
			},
			{
				Title:   "pic6",
				Path:    "/photos/child2/pic6",
				AlbumID: 3,
			},
		}

		verifyResult(t, expected_thumbnails, resp)
	})

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
