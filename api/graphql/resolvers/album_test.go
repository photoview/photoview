package resolvers_test

import (
	"context"
	"testing"

	api "github.com/photoview/photoview/api/graphql"
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/graphql/resolvers"

	"github.com/99designs/gqlgen/client"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/photoview/photoview/api/graphql/auth"
	"github.com/photoview/photoview/api/test_utils"
	"github.com/stretchr/testify/assert"
)

func addContext(ctx context.Context) client.Option {
	return func(bd *client.Request) {
		bd.HTTP = bd.HTTP.WithContext(ctx)
	}
}

func TestAlbumCover(t *testing.T) {
	db := test_utils.DatabaseTest(t)

	rootAlbum := models.Album{
		Title: "root",
		Path:  "/photos",
	}

	if !assert.NoError(t, db.Save(&rootAlbum).Error) {
		return
	}

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
		},
	}

	if !assert.NoError(t, db.Save(&children).Error) {
		return
	}

	photos := []models.Media{
		{
			Title:   "pic1",
			Path:    "/photos/pic1",
			AlbumID: rootAlbum.ID,
		},
		{
			Title:   "pic2",
			Path:    "/photos/pic2",
			AlbumID: rootAlbum.ID,
		},
		{
			Title:   "pic3",
			Path:    "/photos/child1/pic3",
			AlbumID: children[0].ID,
		},
		{
			Title:   "pic4",
			Path:    "/photos/child1/pic4",
			AlbumID: children[0].ID,
		},
		{
			Title:   "pic5",
			Path:    "/photos/child2/pic5",
			AlbumID: children[1].ID,
		},
		{
			Title:   "pic6",
			Path:    "/photos/child2/pic6",
			AlbumID: children[1].ID,
		},
	}

	if !assert.NoError(t, db.Save(&photos).Error) {
		return
	}

	if !assert.NoError(t, db.Model(&children[1]).Update("cover_id", &photos[5].ID).Error) {
		return
	}

	if !assert.NoError(t, db.Model(&children[0]).Update("cover_id", &photos[3].ID).Error) {
		return
	}

	photoUrls := []models.MediaURL{
		{
			MediaID: photos[0].ID,
			Media: &photos[0],
		},
		{
			MediaID: photos[1].ID,
			Media: &photos[1],
		},
		{
			MediaID: photos[2].ID,
			Media: &photos[2],
		},
		{
			MediaID: photos[3].ID,
			Media: &photos[3],
		},
		{
			MediaID: photos[4].ID,
			Media: &photos[4],
		},
		{
			MediaID: photos[5].ID,
			Media: &photos[5],
		},
	}

	if !assert.NoError(t, db.Save(&photoUrls).Error) {
		return
	}

	pass := "<hashed_password>"
	regularUser := models.User{
		Username: "user1",
		Password: &pass,
		Admin:    false,
	}

	if !assert.NoError(t, db.Save(&regularUser).Error) {
		return
	}

	if !assert.NoError(t, db.Model(&regularUser).Association("Albums").Append(&rootAlbum)) {
		return
	}

	if !assert.NoError(t, db.Model(&regularUser).Association("Albums").Append(&children)) {
		return
	}

	ctx := auth.AddUserToContext(context.TODO(), &regularUser)

	c := client.New(handler.NewDefaultServer(api.NewExecutableSchema(api.Config{
		Resolvers: &resolvers.Resolver{
			Database: db,
		},
		Directives: api.DirectiveRoot{
			IsAuthorized: api.IsAuthorized,
		},
	})))

	// Single test since we cannot rely on the tests being performed sequentially
	t.Run("Album get and reset cover photos", func(t *testing.T) {

		var resp struct {
			Album struct {
				Thumbnail struct {
					Title string
				}
			}
		}

		q := `query ($albumID: ID!){ album (id: $albumID) {
			thumbnail {
				title
			}
		}}
		`

		postErr := c.Post(
			q,
			&resp,
			client.Var("albumID", &rootAlbum.ID),
			addContext(ctx),
		)
		if !assert.NoError(t, postErr) {
			return
		}

		// Should return pic1 since no coverID has been set
		assert.EqualValues(t, "pic1", resp.Album.Thumbnail.Title)

		qErr := c.Post(
			q,
			&resp,
			client.Var("albumID", &children[0].ID),
			addContext(ctx),
		)
		if !assert.NoError(t, qErr) {
			return
		}

		// coverID has already been set
		assert.EqualValues(t, "pic4", resp.Album.Thumbnail.Title)

		var resetResp struct {
			ResetAlbumCover struct {
				CoverID int
			}
		}

		m := `mutation resetCover($albumID: ID!) {
	    resetAlbumCover(albumID: $albumID) {
				coverID
			}
	  }
		`
		mErr := c.Post(
			m,
			&resetResp,
			client.Var("albumID", &children[0].ID),
			addContext(ctx),
		)
		if !assert.NoError(t, mErr) {
			return
		}

		assert.EqualValues(t, 0, resetResp.ResetAlbumCover.CoverID)
	})
	
	t.Run("Album change cover photos", func(t *testing.T) {

		var resp struct {
			SetAlbumCover struct {
				CoverID int
			}
		}

		q := `mutation changeCover($coverID: ID!) {
    	setAlbumCover(coverID: $coverID) {
      	coverID,
    	}
	  }
		`

		postErr := c.Post(
			q,
			&resp,
			client.Var("coverID", &photos[4].ID),
			addContext(ctx),
		)
		if !assert.NoError(t, postErr) {
			return
		}

		assert.EqualValues(t, &photos[4].ID, &resp.SetAlbumCover.CoverID)

	})

}
