package cleanup_tasks_test

import (
	"errors"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/otiai10/copy"
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner/face_detection"
	"github.com/photoview/photoview/api/test_utils"
	scanner_utils "github.com/photoview/photoview/api/test_utils/scanner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestCleanupMediaKeepsMediaThatFailedToScan(t *testing.T) {
	test_utils.FilesystemTest(t)
	db := test_utils.DatabaseTest(t)

	require.NoError(t, face_detection.InitializeFaceDetector(db))

	testDir := t.TempDir()
	failingPath := path.Join(testDir, "lilac_lilac_bush_lilac.jpg")
	removedPath := path.Join(testDir, "buttercup_close_summer_yellow.jpg")

	for _, mediaPath := range []string{failingPath, removedPath} {
		require.NoError(t, copy.Copy(path.Join("../../test_media/library", path.Base(mediaPath)), mediaPath))
	}

	pass := "1234"
	user, err := models.RegisterUser(db, "user1", &pass, true)
	require.NoError(t, err)

	rootAlbum := models.Album{
		Title: "root album",
		Path:  testDir,
	}
	require.NoError(t, db.Save(&rootAlbum).Error)
	require.NoError(t, db.Model(user).Association("Albums").Append(&rootAlbum))

	mediaPaths := func() []string {
		var paths []string
		require.NoError(t, db.Model(&models.Media{}).Order("path").Pluck("path", &paths).Error)

		return paths
	}

	scanner_utils.RunScannerOnUser(t, db, user)
	require.Equal(t, []string{removedPath, failingPath}, mediaPaths())

	// Fail the scanner's lookup of one existing file, the way a transient
	// database error would, and remove the other file from disk.
	failingHash := models.MD5Hash(failingPath)
	const callbackName = "test:fail_media_lookup"

	require.NoError(t, db.Callback().Query().After("gorm:query").Register(callbackName, func(tx *gorm.DB) {
		if !strings.Contains(tx.Statement.SQL.String(), "path_hash = ") {
			return
		}

		for _, v := range tx.Statement.Vars {
			if v == failingHash {
				tx.AddError(errors.New("simulated database error"))

				return
			}
		}
	}))
	t.Cleanup(func() {
		_ = db.Callback().Query().Remove(callbackName)
	})

	require.NoError(t, os.Remove(removedPath))
	scanner_utils.RunScannerOnUser(t, db, user)

	// The file that failed to scan is still on disk, so its media is kept, while
	// the removed file is cleaned up in the same scan.
	assert.Equal(t, []string{failingPath}, mediaPaths())
}

// TestScanRetriesATransientDatabaseFailure covers the step before the fallback
// above: a file whose database transaction fails once is retried, so the scan
// that found it also stores it. Without the retry the file would end up as
// unscanned, and a new file has no media row to keep - the fallback cannot
// help it, and the photo would be missing until the next scan.
func TestScanRetriesATransientDatabaseFailure(t *testing.T) {
	test_utils.FilesystemTest(t)
	db := test_utils.DatabaseTest(t)

	require.NoError(t, face_detection.InitializeFaceDetector(db))

	testDir := t.TempDir()
	mediaPath := path.Join(testDir, "lilac_lilac_bush_lilac.jpg")
	require.NoError(t, copy.Copy(path.Join("../../test_media/library", path.Base(mediaPath)), mediaPath))

	pass := "1234"
	user, err := models.RegisterUser(db, "retry_user", &pass, true)
	require.NoError(t, err)

	rootAlbum := models.Album{Title: "root album", Path: testDir}
	require.NoError(t, db.Save(&rootAlbum).Error)
	require.NoError(t, db.Model(user).Association("Albums").Append(&rootAlbum))

	// Fail the lookup of that file exactly once, the way a lock timeout or a
	// dropped connection would.
	mediaHash := models.MD5Hash(mediaPath)
	const callbackName = "test:fail_media_lookup_once"
	failures := 0

	require.NoError(t, db.Callback().Query().After("gorm:query").Register(callbackName, func(tx *gorm.DB) {
		if failures > 0 || !strings.Contains(tx.Statement.SQL.String(), "path_hash = ") {
			return
		}

		for _, v := range tx.Statement.Vars {
			if v == mediaHash {
				failures++
				tx.AddError(errors.New("simulated transient database error"))

				return
			}
		}
	}))
	t.Cleanup(func() {
		_ = db.Callback().Query().Remove(callbackName)
	})

	scanner_utils.RunScannerOnUser(t, db, user)

	assert.Equal(t, 1, failures, "the test's failure must have been injected")

	var paths []string
	require.NoError(t, db.Model(&models.Media{}).Order("path").Pluck("path", &paths).Error)
	assert.Equal(t, []string{mediaPath}, paths, "the retried file belongs in the library")
}
