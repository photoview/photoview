package routes

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/graphql/models/actions"
	"github.com/photoview/photoview/api/test_utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthenticateAlbumShareTokenFinishesOnCyclicParents(t *testing.T) {
	db := test_utils.DatabaseTest(t)

	user, err := models.RegisterUser(db, "username", nil, false)
	require.NoError(t, err)

	// Two albums whose parents point at each other. The scanner never creates
	// this, but a hand-edited or corrupted database can hold it.
	first := models.Album{Title: "first", Path: "/cycle/first"}
	require.NoError(t, db.Model(&user).Association("Albums").Append(&first))

	second := models.Album{Title: "second", Path: "/cycle/second", ParentAlbumID: &first.ID}
	require.NoError(t, db.Save(&second).Error)
	require.NoError(t, db.Model(&first).Update("parent_album_id", second.ID).Error)

	unrelated := models.Album{Title: "unrelated", Path: "/unrelated"}
	require.NoError(t, db.Save(&unrelated).Error)

	shareToken, err := actions.AddAlbumShare(db, user, first.ID, nil, nil, nil)
	require.NoError(t, err)

	authenticate := func(t *testing.T, album *models.Album) (bool, int, error) {
		// A query that never ends fails the test instead of hanging the suite.
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		t.Cleanup(cancel)

		url := fmt.Sprintf("/download/album/%d?token=%s", album.ID, shareToken.Value)
		req := httptest.NewRequest("GET", url, nil)

		success, _, status, err := authenticateAlbum(album, db.WithContext(ctx), req)

		return success, status, err
	}

	t.Run("Album inside the cycle is shared", func(t *testing.T) {
		success, status, err := authenticate(t, &second)
		require.NoError(t, err)
		assert.True(t, success)
		assert.Equal(t, http.StatusAccepted, status)
	})

	t.Run("Album outside the cycle is refused", func(t *testing.T) {
		success, status, err := authenticate(t, &unrelated)
		assert.EqualError(t, err, "invalid share token")
		assert.False(t, success)
		assert.Equal(t, http.StatusForbidden, status)
	})
}
