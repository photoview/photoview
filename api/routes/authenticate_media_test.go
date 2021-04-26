package routes

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/photoview/photoview/api/graphql/auth"
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/graphql/models/actions"
	"github.com/photoview/photoview/api/test_utils"
	"github.com/stretchr/testify/assert"
)

func TestAuthenticateMedia(t *testing.T) {
	db := test_utils.DatabaseTest(t)

	user, err := models.RegisterUser(db, "username", nil, false)
	if !assert.NoError(t, err) {
		return
	}

	album := models.Album{
		Title: "my_album",
	}

	if !assert.NoError(t, db.Model(&user).Association("Albums").Append(&album)) {
		return
	}

	media := models.Media{
		Title:   "my_media",
		Path:    "/photos/image.jpg",
		AlbumID: album.ID,
	}

	if !assert.NoError(t, db.Save(&media).Error) {
		return
	}

	t.Run("Authorized request", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/photo/image.jpg", strings.NewReader("IMAGE DATA"))
		ctx := auth.AddUserToContext(req.Context(), user)
		req = req.WithContext(ctx)

		success, responseMessage, responseStatus, err := authenticateMedia(&media, db, req)

		assert.NoError(t, err)
		assert.True(t, success)
		assert.Equal(t, responseMessage, "success")
		assert.Equal(t, responseStatus, http.StatusAccepted)
	})

	t.Run("Request without access token", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/photo/image.jpg", strings.NewReader("IMAGE DATA"))

		success, responseMessage, responseStatus, err := authenticateMedia(&media, db, req)

		assert.NoError(t, err)
		assert.False(t, success)
		assert.Equal(t, responseMessage, "unauthorized")
		assert.Equal(t, responseStatus, http.StatusForbidden)
	})

	t.Run("Request without access token", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/photo/image.jpg", strings.NewReader("IMAGE DATA"))

		success, responseMessage, responseStatus, err := authenticateMedia(&media, db, req)

		assert.NoError(t, err)
		assert.False(t, success)
		assert.Equal(t, responseMessage, "unauthorized")
		assert.Equal(t, responseStatus, http.StatusForbidden)
	})

	expire := time.Now().Add(time.Hour * 24 * 30)
	tokenPassword := "token-password-123"
	shareToken, err := actions.AddMediaShare(db, user.ID, media.ID, &expire, &tokenPassword)
	if !assert.NoError(t, err) {
		return
	}

	t.Run("Request with share token", func(t *testing.T) {
		url := fmt.Sprintf("/photo/image.jpg?token=%s", shareToken.Value)
		req := httptest.NewRequest("GET", url, strings.NewReader("IMAGE DATA"))

		cookie := http.Cookie{
			Name:  fmt.Sprintf("share-token-pw-%s", shareToken.Value),
			Value: tokenPassword,
		}
		req.AddCookie(&cookie)

		success, responseMessage, responseStatus, err := authenticateMedia(&media, db, req)

		assert.NoError(t, err)
		assert.True(t, success)
		assert.Equal(t, responseMessage, "success")
		assert.Equal(t, responseStatus, http.StatusAccepted)
	})

}
