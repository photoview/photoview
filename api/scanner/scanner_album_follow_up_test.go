package scanner

import (
	"context"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner/scanner_cache"
	"github.com/photoview/photoview/api/scanner/scanner_task"
	"github.com/photoview/photoview/api/test_utils"
)

func TestScanMediaWithFollowUpRetriesNewMediaAfterFailure(t *testing.T) {
	db := test_utils.DatabaseTest(t)
	album := models.Album{Title: "album", Path: test_utils.PathFromAPIRoot("scanner", "test_media", "library")}
	require.NoError(t, db.Create(&album).Error)

	mediaPath := test_utils.PathFromAPIRoot("scanner", "test_media", "real_media", "jpeg.jpg")
	ctx := scanner_task.NewTaskContext(context.Background(), db, &album, scanner_cache.MakeAlbumCache())

	media, err := scanMediaWithFollowUp(ctx, mediaPath, func(scanner_task.TaskContext, *models.Media, bool) error {
		return errors.New("follow-up failed")
	})
	require.Error(t, err)
	assert.Nil(t, media)

	var count int64
	require.NoError(t, db.Model(&models.Media{}).Where("path_hash = ?", models.MD5Hash(mediaPath)).Count(&count).Error)
	assert.Zero(t, count)

	var retriedAsNewMedia bool
	media, err = scanMediaWithFollowUp(ctx, mediaPath, func(_ scanner_task.TaskContext, _ *models.Media, isNewMedia bool) error {
		retriedAsNewMedia = isNewMedia
		return nil
	})
	require.NoError(t, err)
	require.NotNil(t, media)
	assert.True(t, retriedAsNewMedia)
}
