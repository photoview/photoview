package routes

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/kkovaletp/photoview/api/graphql/auth"
	"github.com/kkovaletp/photoview/api/graphql/models"
	"github.com/kkovaletp/photoview/api/graphql/models/actions"
	"github.com/kkovaletp/photoview/api/test_utils"
	"github.com/stretchr/testify/assert"
)

func TestAuthenticateRoute(t *testing.T) {
	const imageData = "IMAGE DATA"
	const albumData = "ALBUM DATA"

	db := test_utils.DatabaseTest(t)

	user, err := models.RegisterUser(db, "username", nil, false)
	if !assert.NoError(t, err) {
		return
	}

	album := models.Album{
		Title: "my_album",
		Path:  "/photos",
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

	t.Run("Authenticate Media", func(t *testing.T) {
		t.Run("Authorized request", func(t *testing.T) {
			req := httptest.NewRequest("GET", "/photo/image.jpg", strings.NewReader(imageData))
			ctx := auth.AddUserToContext(req.Context(), user)
			req = req.WithContext(ctx)

			success, responseMessage, responseStatus, err := authenticateMedia(&media, db, req)

			assert.NoError(t, err)
			assert.True(t, success)
			assert.Equal(t, responseMessage, "success")
			assert.Equal(t, responseStatus, http.StatusAccepted)
		})

		t.Run("Request without access token", func(t *testing.T) {
			req := httptest.NewRequest("GET", "/photo/image.jpg", strings.NewReader(imageData))

			success, responseMessage, responseStatus, err := authenticateMedia(&media, db, req)

			assert.Error(t, err)
			assert.False(t, success)
			assert.Equal(t, responseMessage, "unauthorized")
			assert.Equal(t, responseStatus, http.StatusForbidden)
		})

		expire := time.Now().Add(time.Hour * 24 * 30)
		tokenPassword := "token-password-123"
		shareToken, err := actions.AddMediaShare(db, user, media.ID, &expire, &tokenPassword)
		if !assert.NoError(t, err) {
			return
		}

		t.Run("Request with share token", func(t *testing.T) {
			url := fmt.Sprintf("/photo/image.jpg?token=%s", shareToken.Value)
			req := httptest.NewRequest("GET", url, strings.NewReader(imageData))

			cookie := http.Cookie{
				Name:  fmt.Sprintf("share-token-pw-%s", shareToken.Value),
				Value: tokenPassword,
			}
			req.AddCookie(&cookie)

			success, responseMessage, responseStatus, err := authenticateMedia(&media, db, req)

			assert.NoError(t, err)
			assert.True(t, success)
			assert.Equal(t, "success", responseMessage)
			assert.Equal(t, http.StatusAccepted, responseStatus)
		})

		t.Run("Request with invalid share token", func(t *testing.T) {
			url := fmt.Sprintf("/photo/image.jpg?token=%s", "invalid-token")
			req := httptest.NewRequest("GET", url, strings.NewReader(imageData))
			// Even if a cookie is sent, the token is invalid
			cookie := http.Cookie{
				Name:  fmt.Sprintf("share-token-pw-%s", "invalid-token"),
				Value: "whatever",
			}
			req.AddCookie(&cookie)
			success, responseMessage, responseStatus, err := authenticateMedia(&media, db, req)
			assert.Error(t, err)
			assert.False(t, success)
			assert.Equal(t, "unauthorized", responseMessage)
			assert.Equal(t, http.StatusForbidden, responseStatus)
		})

		t.Run("Request with share token but no password cookie", func(t *testing.T) {
			shareToken, err := actions.AddMediaShare(db, user, media.ID, &expire, &tokenPassword)
			assert.NoError(t, err)
			url := fmt.Sprintf("/photo/image.jpg?token=%s", shareToken.Value)
			req := httptest.NewRequest("GET", url, strings.NewReader(imageData))
			// No cookie provided
			success, responseMessage, responseStatus, err := authenticateMedia(&media, db, req)
			assert.Error(t, err)
			assert.False(t, success)
			assert.Equal(t, "unauthorized", responseMessage)
			assert.Equal(t, http.StatusForbidden, responseStatus)
		})

		t.Run("Request with share token and wrong password", func(t *testing.T) {
			shareToken, err := actions.AddMediaShare(db, user, media.ID, &expire, &tokenPassword)
			assert.NoError(t, err)
			url := fmt.Sprintf("/photo/image.jpg?token=%s", shareToken.Value)
			req := httptest.NewRequest("GET", url, strings.NewReader(imageData))
			cookie := http.Cookie{
				Name:  fmt.Sprintf("share-token-pw-%s", shareToken.Value),
				Value: "incorrect-password",
			}
			req.AddCookie(&cookie)
			success, responseMessage, responseStatus, err := authenticateMedia(&media, db, req)
			assert.Error(t, err)
			assert.False(t, success)
			assert.Equal(t, "unauthorized", responseMessage)
			assert.Equal(t, http.StatusForbidden, responseStatus)
		})

		t.Run("Request with expired share token", func(t *testing.T) {
			expired := time.Now().Add(-time.Hour)
			shareToken, err := actions.AddMediaShare(db, user, media.ID, &expired, &tokenPassword)
			assert.NoError(t, err)
			url := fmt.Sprintf("/photo/image.jpg?token=%s", shareToken.Value)
			req := httptest.NewRequest("GET", url, strings.NewReader(imageData))
			cookie := http.Cookie{
				Name:  fmt.Sprintf("share-token-pw-%s", shareToken.Value),
				Value: tokenPassword,
			}
			req.AddCookie(&cookie)
			success, responseMessage, responseStatus, err := authenticateMedia(&media, db, req)
			assert.Error(t, err)
			assert.False(t, success)
			assert.Equal(t, "unauthorized", responseMessage)
			assert.Equal(t, http.StatusForbidden, responseStatus)
		})
	})

	t.Run("Authenticate Album", func(t *testing.T) {
		t.Run("Authorized request", func(t *testing.T) {
			req := httptest.NewRequest("GET", "/download/album/1", strings.NewReader(albumData))
			ctx := auth.AddUserToContext(req.Context(), user)
			req = req.WithContext(ctx)

			success, responseMessage, responseStatus, err := authenticateAlbum(&album, db, req)

			assert.NoError(t, err)
			assert.True(t, success)
			assert.Equal(t, "success", responseMessage)
			assert.Equal(t, http.StatusAccepted, responseStatus)
		})

		t.Run("Request without access token", func(t *testing.T) {
			req := httptest.NewRequest("GET", "/download/album/1", strings.NewReader(albumData))

			success, responseMessage, responseStatus, err := authenticateAlbum(&album, db, req)

			assert.Error(t, err)
			assert.False(t, success)
			assert.Equal(t, "unauthorized", responseMessage)
			assert.Equal(t, http.StatusForbidden, responseStatus)
		})

		expire := time.Now().Add(time.Hour * 24 * 30)
		tokenPassword := "token-password-123"
		shareToken, err := actions.AddAlbumShare(db, user, album.ID, &expire, &tokenPassword)
		if !assert.NoError(t, err) {
			return
		}

		t.Run("Request with share token", func(t *testing.T) {
			url := fmt.Sprintf("/download/album/1?token=%s", shareToken.Value)
			req := httptest.NewRequest("GET", url, strings.NewReader(albumData))

			cookie := http.Cookie{
				Name:  fmt.Sprintf("share-token-pw-%s", shareToken.Value),
				Value: tokenPassword,
			}
			req.AddCookie(&cookie)

			success, responseMessage, responseStatus, err := authenticateAlbum(&album, db, req)

			assert.NoError(t, err)
			assert.True(t, success)
			assert.Equal(t, "success", responseMessage)
			assert.Equal(t, http.StatusAccepted, responseStatus)
		})

		t.Run("Request with invalid album share token", func(t *testing.T) {
			url := fmt.Sprintf("/download/album/1?token=%s", "invalid-token")
			req := httptest.NewRequest("GET", url, strings.NewReader(albumData))
			cookie := http.Cookie{
				Name:  fmt.Sprintf("share-token-pw-%s", "invalid-token"),
				Value: "whatever",
			}
			req.AddCookie(&cookie)
			success, responseMessage, responseStatus, err := authenticateAlbum(&album, db, req)
			assert.Error(t, err)
			assert.False(t, success)
			assert.Equal(t, "unauthorized", responseMessage)
			assert.Equal(t, http.StatusForbidden, responseStatus)
		})

		t.Run("Request with album share token but no password cookie", func(t *testing.T) {
			shareToken, err := actions.AddAlbumShare(db, user, album.ID, &expire, &tokenPassword)
			assert.NoError(t, err)
			url := fmt.Sprintf("/download/album/1?token=%s", shareToken.Value)
			req := httptest.NewRequest("GET", url, strings.NewReader(albumData))
			// No cookie provided
			success, responseMessage, responseStatus, err := authenticateAlbum(&album, db, req)
			assert.Error(t, err)
			assert.False(t, success)
			assert.Equal(t, "unauthorized", responseMessage)
			assert.Equal(t, http.StatusForbidden, responseStatus)
		})

		t.Run("Request with album share token and wrong password", func(t *testing.T) {
			shareToken, err := actions.AddAlbumShare(db, user, album.ID, &expire, &tokenPassword)
			assert.NoError(t, err)
			url := fmt.Sprintf("/download/album/1?token=%s", shareToken.Value)
			req := httptest.NewRequest("GET", url, strings.NewReader(albumData))
			cookie := http.Cookie{
				Name:  fmt.Sprintf("share-token-pw-%s", shareToken.Value),
				Value: "incorrect-password",
			}
			req.AddCookie(&cookie)
			success, responseMessage, responseStatus, err := authenticateAlbum(&album, db, req)
			assert.Error(t, err)
			assert.False(t, success)
			assert.Equal(t, "unauthorized", responseMessage)
			assert.Equal(t, http.StatusForbidden, responseStatus)
		})

		t.Run("Request with expired album share token", func(t *testing.T) {
			expired := time.Now().Add(-time.Hour)
			shareToken, err := actions.AddAlbumShare(db, user, album.ID, &expired, &tokenPassword)
			assert.NoError(t, err)
			url := fmt.Sprintf("/download/album/1?token=%s", shareToken.Value)
			req := httptest.NewRequest("GET", url, strings.NewReader(albumData))
			cookie := http.Cookie{
				Name:  fmt.Sprintf("share-token-pw-%s", shareToken.Value),
				Value: tokenPassword,
			}
			req.AddCookie(&cookie)
			success, responseMessage, responseStatus, err := authenticateAlbum(&album, db, req)
			assert.Error(t, err)
			assert.False(t, success)
			assert.Equal(t, "unauthorized", responseMessage)
			assert.Equal(t, http.StatusForbidden, responseStatus)
		})
	})
}
