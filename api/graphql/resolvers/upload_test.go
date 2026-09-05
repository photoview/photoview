package resolvers

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/photoview/photoview/api/graphql/auth"
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/test_utils"
	"github.com/stretchr/testify/assert"
)

func setupUploadMutationTest(t *testing.T) (*mutationResolver, *models.User, *models.User) {
	t.Helper()

	test_utils.FilesystemTest(t)
	db := test_utils.DatabaseTest(t)

	// Stub out the real scanner queue: it starts a background worker
	// goroutine on first use and isn't needed to verify the resolvers'
	// own DB/filesystem behavior.
	origAddAlbumToQueue := addAlbumToQueue
	addAlbumToQueue = func(album *models.Album) error { return nil }
	t.Cleanup(func() { addAlbumToQueue = origAddAlbumToQueue })

	uploader, err := models.RegisterUser(db, "uploader", nil, false)
	assert.NoError(t, err)
	uploader.CanUpload = true
	assert.NoError(t, db.Save(uploader).Error)

	nonUploader, err := models.RegisterUser(db, "non_uploader", nil, false)
	assert.NoError(t, err)

	r := &mutationResolver{Resolver: &Resolver{database: db}}

	return r, uploader, nonUploader
}

func makeTestRootAlbum(t *testing.T, r *mutationResolver, owner *models.User, title string) *models.Album {
	t.Helper()

	albumPath := filepath.Join(t.TempDir(), title)
	assert.NoError(t, os.Mkdir(albumPath, 0o755))

	album := models.Album{Title: title, Path: albumPath}
	assert.NoError(t, r.database.Save(&album).Error)
	assert.NoError(t, r.database.Model(owner).Association("Albums").Append(&album))

	return &album
}

func TestCreateAlbumFolder(t *testing.T) {
	r, uploader, nonUploader := setupUploadMutationTest(t)
	root := makeTestRootAlbum(t, r, uploader, "root")
	ctx := auth.AddUserToContext(context.Background(), uploader)

	t.Run("happy path", func(t *testing.T) {
		album, err := r.CreateAlbumFolder(ctx, root.ID, "vacation")
		assert.NoError(t, err)
		if assert.NotNil(t, album) {
			assert.Equal(t, "vacation", album.Title)
			assert.Equal(t, filepath.Join(root.Path, "vacation"), album.Path)

			info, statErr := os.Stat(album.Path)
			assert.NoError(t, statErr)
			assert.True(t, info.IsDir())

			owns, err := uploader.OwnsAlbum(r.database, album)
			assert.NoError(t, err)
			assert.True(t, owns)
		}
	})

	t.Run("name collision", func(t *testing.T) {
		_, err := r.CreateAlbumFolder(ctx, root.ID, "vacation")
		assert.Error(t, err)
	})

	t.Run("path traversal in name is rejected", func(t *testing.T) {
		_, err := r.CreateAlbumFolder(ctx, root.ID, "../evil")
		assert.Error(t, err)
		_, statErr := os.Stat(filepath.Join(filepath.Dir(root.Path), "evil"))
		assert.True(t, os.IsNotExist(statErr))
	})

	t.Run("permission denied for non-uploader", func(t *testing.T) {
		nonUploaderCtx := auth.AddUserToContext(context.Background(), nonUploader)
		_, err := r.CreateAlbumFolder(nonUploaderCtx, root.ID, "denied")
		assert.Error(t, err)
	})
}

func TestDeleteAlbum(t *testing.T) {
	r, uploader, nonUploader := setupUploadMutationTest(t)
	root := makeTestRootAlbum(t, r, uploader, "root")
	ctx := auth.AddUserToContext(context.Background(), uploader)

	t.Run("root album cannot be deleted", func(t *testing.T) {
		_, err := r.DeleteAlbum(ctx, root.ID)
		assert.Error(t, err)
	})

	t.Run("happy path moves the folder to trash and removes the rows", func(t *testing.T) {
		folder, err := r.CreateAlbumFolder(ctx, root.ID, "to_delete")
		assert.NoError(t, err)
		originalPath := folder.Path

		sub, err := r.CreateAlbumFolder(ctx, folder.ID, "nested")
		assert.NoError(t, err)

		ok, err := r.DeleteAlbum(ctx, folder.ID)
		assert.NoError(t, err)
		assert.True(t, ok)

		_, statErr := os.Stat(originalPath)
		assert.True(t, os.IsNotExist(statErr), "original folder should no longer exist")

		trashEntries, err := os.ReadDir(filepath.Join(root.Path, ".trash"))
		assert.NoError(t, err)
		assert.Len(t, trashEntries, 1)

		assert.Error(t, r.database.First(&models.Album{}, folder.ID).Error)
		assert.Error(t, r.database.First(&models.Album{}, sub.ID).Error)
	})

	t.Run("permission denied for non-uploader", func(t *testing.T) {
		folder, err := r.CreateAlbumFolder(ctx, root.ID, "protected")
		assert.NoError(t, err)

		nonUploaderCtx := auth.AddUserToContext(context.Background(), nonUploader)
		_, err = r.DeleteAlbum(nonUploaderCtx, folder.ID)
		assert.Error(t, err)

		_, statErr := os.Stat(folder.Path)
		assert.NoError(t, statErr, "folder should be untouched after a denied delete")
	})
}

func TestMoveAlbum(t *testing.T) {
	r, uploader, nonUploader := setupUploadMutationTest(t)
	root := makeTestRootAlbum(t, r, uploader, "root")
	ctx := auth.AddUserToContext(context.Background(), uploader)

	t.Run("root album cannot be moved", func(t *testing.T) {
		destination, err := r.CreateAlbumFolder(ctx, root.ID, "destination1")
		assert.NoError(t, err)

		_, err = r.MoveAlbum(ctx, root.ID, destination.ID)
		assert.Error(t, err)
	})

	t.Run("happy path preserves media and moves on disk", func(t *testing.T) {
		source, err := r.CreateAlbumFolder(ctx, root.ID, "source")
		assert.NoError(t, err)
		destination, err := r.CreateAlbumFolder(ctx, root.ID, "destination2")
		assert.NoError(t, err)

		media := models.Media{
			Title:   "photo.jpg",
			Path:    filepath.Join(source.Path, "photo.jpg"),
			AlbumID: source.ID,
		}
		assert.NoError(t, r.database.Save(&media).Error)

		moved, err := r.MoveAlbum(ctx, source.ID, destination.ID)
		assert.NoError(t, err)
		if !assert.NotNil(t, moved) {
			return
		}

		wantPath := filepath.Join(destination.Path, "source")
		assert.Equal(t, wantPath, moved.Path)
		assert.Equal(t, destination.ID, *moved.ParentAlbumID)

		_, statErr := os.Stat(wantPath)
		assert.NoError(t, statErr)
		_, statErr = os.Stat(source.Path)
		assert.True(t, os.IsNotExist(statErr))

		var updatedMedia models.Media
		assert.NoError(t, r.database.First(&updatedMedia, media.ID).Error)
		assert.Equal(t, filepath.Join(wantPath, "photo.jpg"), updatedMedia.Path)
	})

	t.Run("cannot move an album into its own descendant", func(t *testing.T) {
		parent, err := r.CreateAlbumFolder(ctx, root.ID, "cycle_parent")
		assert.NoError(t, err)
		child, err := r.CreateAlbumFolder(ctx, parent.ID, "cycle_child")
		assert.NoError(t, err)

		_, err = r.MoveAlbum(ctx, parent.ID, child.ID)
		assert.Error(t, err)
	})

	t.Run("permission denied for non-uploader", func(t *testing.T) {
		source, err := r.CreateAlbumFolder(ctx, root.ID, "source_denied")
		assert.NoError(t, err)
		destination, err := r.CreateAlbumFolder(ctx, root.ID, "destination_denied")
		assert.NoError(t, err)

		nonUploaderCtx := auth.AddUserToContext(context.Background(), nonUploader)
		_, err = r.MoveAlbum(nonUploaderCtx, source.ID, destination.ID)
		assert.Error(t, err)
	})
}
