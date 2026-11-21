package auth_test

import (
	"context"
	"strings"
	"testing"

	"github.com/photoview/photoview/api/graphql/auth"
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/test_utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// People fields used to copy and sort the caller's album IDs on every resolver.
// One operation keeps a single sorted slice and does not query membership again.
func TestUserAlbumIDsReusedWithinOperation(t *testing.T) {
	db := test_utils.DatabaseTest(t)
	password := "1234"
	user, err := models.RegisterUser(db, "album-ids", &password, false)
	require.NoError(t, err)

	for _, path := range []string{"/photos/a", "/photos/b", "/photos/c"} {
		album := models.Album{Title: path, Path: path}
		require.NoError(t, db.Save(&album).Error)
		require.NoError(t, db.Model(user).Association("Albums").Append(&album))
	}

	ctx := auth.WithAlbumAccessCache(context.Background())
	var queries int
	const callbackName = "test:album-id-reuse"
	_ = db.Callback().Query().Remove(callbackName)
	require.NoError(t, db.Callback().Query().After("gorm:query").Register(callbackName, func(tx *gorm.DB) {
		if strings.Contains(strings.ToLower(tx.Statement.SQL.String()), "user_albums") {
			queries++
		}
	}))
	t.Cleanup(func() { _ = db.Callback().Query().Remove(callbackName) })

	first, err := auth.UserAlbumIDs(ctx, db, user)
	require.NoError(t, err)
	require.Len(t, first, 3)
	assert.Less(t, first[0], first[1])
	assert.Less(t, first[1], first[2])

	second, err := auth.UserAlbumIDs(ctx, db, user)
	require.NoError(t, err)
	assert.Equal(t, first, second)
	require.NotEmpty(t, first)
	assert.Same(t, &first[0], &second[0], "the operation must reuse the sorted album ID slice")
	assert.Equal(t, 1, queries)

	// A context without the operation cache reads the database again.
	third, err := auth.UserAlbumIDs(context.Background(), db, user)
	require.NoError(t, err)
	assert.Equal(t, first, third)
	assert.Equal(t, 2, queries)
}
