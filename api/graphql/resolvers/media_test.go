package resolvers

import (
	"context"
	"testing"

	"github.com/photoview/photoview/api/graphql/auth"
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/graphql/models/actions"
	"github.com/photoview/photoview/api/test_utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type mediaTokenFixtures struct {
	resolver    *queryResolver
	albumToken  *models.ShareToken
	mediaToken  *models.ShareToken
	otherToken  *models.ShareToken
	media       *models.Media
	otherMedia  *models.Media
	album       *models.Album
	otherAlbum  *models.Album
	user        *models.User
	otherPerson *models.User
}

func setupMediaTokenFixtures(t *testing.T) mediaTokenFixtures {
	t.Helper()

	db := test_utils.DatabaseTest(t)

	password := "1234"
	user, err := models.RegisterUser(db, "token_owner", &password, false)
	require.NoError(t, err)

	otherUser, err := models.RegisterUser(db, "other_user", &password, false)
	require.NoError(t, err)

	album := models.Album{Title: "shared album", Path: "/photos/shared"}
	require.NoError(t, db.Save(&album).Error)
	require.NoError(t, db.Model(&user).Association("Albums").Append(&album))

	otherAlbum := models.Album{Title: "other album", Path: "/photos/other"}
	require.NoError(t, db.Save(&otherAlbum).Error)
	require.NoError(t, db.Model(&otherUser).Association("Albums").Append(&otherAlbum))

	media := models.Media{Title: "shared photo", Path: "/photos/shared/photo.jpg", AlbumID: album.ID}
	require.NoError(t, db.Save(&media).Error)

	// The owner path only returns media that has at least one URL row.
	mediaURL := models.MediaURL{
		MediaID:     media.ID,
		MediaName:   "photo.jpg",
		Width:       100,
		Height:      100,
		Purpose:     models.MediaOriginal,
		ContentType: "image/jpeg",
		FileSize:    1024,
	}
	require.NoError(t, db.Save(&mediaURL).Error)

	otherMedia := models.Media{Title: "other photo", Path: "/photos/other/photo.jpg", AlbumID: otherAlbum.ID}
	require.NoError(t, db.Save(&otherMedia).Error)

	albumToken, err := actions.AddAlbumShare(db, user, album.ID, nil, nil, nil)
	require.NoError(t, err)

	mediaToken, err := actions.AddMediaShare(db, user, media.ID, nil, nil, nil)
	require.NoError(t, err)

	otherToken, err := actions.AddMediaShare(db, otherUser, otherMedia.ID, nil, nil, nil)
	require.NoError(t, err)

	return mediaTokenFixtures{
		resolver:    &queryResolver{Resolver: &Resolver{database: db}},
		albumToken:  albumToken,
		mediaToken:  mediaToken,
		otherToken:  otherToken,
		media:       &media,
		otherMedia:  &otherMedia,
		album:       &album,
		otherAlbum:  &otherAlbum,
		user:        user,
		otherPerson: otherUser,
	}
}

func tokenCredentials(token *models.ShareToken) *models.ShareTokenCredentials {
	return &models.ShareTokenCredentials{Token: token.Value}
}

// An album share token has no media ID. Supplying one to the media query must be
// rejected with a controlled error rather than dereferencing the nil MediaID.
func TestMediaQueryRejectsAlbumShareToken(t *testing.T) {
	fixtures := setupMediaTokenFixtures(t)

	media, err := fixtures.resolver.Media(context.Background(), fixtures.media.ID, tokenCredentials(fixtures.albumToken))

	assert.ErrorIs(t, err, auth.ErrUnauthorized)
	assert.Nil(t, media)
}

// The same credential must not be enough to read a media item in the album either,
// so the rejection cannot depend on which media ID was asked for.
func TestMediaQueryRejectsAlbumShareTokenForItsOwnAlbum(t *testing.T) {
	fixtures := setupMediaTokenFixtures(t)

	media, err := fixtures.resolver.Media(context.Background(), fixtures.otherMedia.ID, tokenCredentials(fixtures.albumToken))

	assert.ErrorIs(t, err, auth.ErrUnauthorized)
	assert.Nil(t, media)
}

func TestMediaQueryAcceptsMatchingMediaShareToken(t *testing.T) {
	fixtures := setupMediaTokenFixtures(t)

	media, err := fixtures.resolver.Media(context.Background(), fixtures.media.ID, tokenCredentials(fixtures.mediaToken))

	require.NoError(t, err)
	require.NotNil(t, media)
	assert.Equal(t, fixtures.media.ID, media.ID)
}

// A media token for one media item must not open another one.
func TestMediaQueryRejectsMismatchedMediaShareToken(t *testing.T) {
	fixtures := setupMediaTokenFixtures(t)

	media, err := fixtures.resolver.Media(context.Background(), fixtures.otherMedia.ID, tokenCredentials(fixtures.mediaToken))

	assert.ErrorIs(t, err, auth.ErrUnauthorized)
	assert.Nil(t, media)
}

func TestMediaQueryRequiresCredentials(t *testing.T) {
	fixtures := setupMediaTokenFixtures(t)

	media, err := fixtures.resolver.Media(context.Background(), fixtures.media.ID, nil)

	assert.ErrorIs(t, err, auth.ErrUnauthorized)
	assert.Nil(t, media)
}

func TestMediaQueryRejectsUnknownShareToken(t *testing.T) {
	fixtures := setupMediaTokenFixtures(t)

	credentials := &models.ShareTokenCredentials{Token: "not-a-real-token"}
	media, err := fixtures.resolver.Media(context.Background(), fixtures.media.ID, credentials)

	assert.EqualError(t, err, "share not found")
	assert.Nil(t, media)
}

// The owner reaches the same media with their own session and no share token.
func TestMediaQueryAcceptsOwnerSession(t *testing.T) {
	fixtures := setupMediaTokenFixtures(t)

	media, err := fixtures.resolver.Media(
		context.Background(),
		fixtures.media.ID,
		nil,
	)

	assert.ErrorIs(t, err, auth.ErrUnauthorized)
	assert.Nil(t, media)

	ctx := auth.AddUserToContext(context.Background(), fixtures.user)
	media, err = fixtures.resolver.Media(ctx, fixtures.media.ID, nil)

	require.NoError(t, err)
	require.NotNil(t, media)
	assert.Equal(t, fixtures.media.ID, media.ID)
}

// The credential is rejected on its type alone, so it is not silently ignored when a
// valid session is also present: the caller is told the credential does not apply and
// can retry without it.
func TestMediaQueryRejectsAlbumShareTokenForLoggedInOwner(t *testing.T) {
	fixtures := setupMediaTokenFixtures(t)

	ctx := auth.AddUserToContext(context.Background(), fixtures.user)
	token := tokenCredentials(fixtures.albumToken)

	media, err := fixtures.resolver.Media(ctx, fixtures.media.ID, token)

	assert.ErrorIs(t, err, auth.ErrUnauthorized)
	assert.Nil(t, media)

	// The same request without the inapplicable credential still resolves.
	media, err = fixtures.resolver.Media(ctx, fixtures.media.ID, nil)
	require.NoError(t, err)
	require.NotNil(t, media)
	assert.Equal(t, fixtures.media.ID, media.ID)
}

// A media token that outlives its media row must not be dereferenced into a panic
// either: deleting the media cascades the token away, and the lookup then misses.
func TestMediaQueryRejectsMediaTokenAfterMediaDeletion(t *testing.T) {
	fixtures := setupMediaTokenFixtures(t)

	require.NoError(t, fixtures.resolver.DB(context.Background()).Transaction(func(tx *gorm.DB) error {
		return tx.Delete(&models.Media{}, fixtures.media.ID).Error
	}))

	media, err := fixtures.resolver.Media(context.Background(), fixtures.media.ID, tokenCredentials(fixtures.mediaToken))

	assert.EqualError(t, err, "share not found")
	assert.Nil(t, media)
}
