package routes

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/kkovaletp/photoview/api/graphql/auth"
	"github.com/kkovaletp/photoview/api/graphql/models"
	"github.com/kkovaletp/photoview/api/scanner"
	"github.com/kkovaletp/photoview/api/test_utils"
	"github.com/kkovaletp/photoview/api/utils"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestPhotoRoutes(t *testing.T) {
	db := test_utils.DatabaseTest(t)

	user, err := models.RegisterUser(db, "testuser", nil, false)
	assert.NoError(t, err)
	album := models.Album{Title: "test_album", Path: "/photos"}
	assert.NoError(t, db.Model(&user).Association("Albums").Append(&album))

	media := models.Media{
		Title:    "test_media",
		Path:     "/photos/test_image.jpg",
		AlbumID:  album.ID,
		DateShot: time.Now(),
		Type:     models.MediaTypePhoto,
	}
	assert.NoError(t, db.Save(&media).Error)

	mediaURL := models.MediaURL{
		MediaID:     media.ID,
		Media:       &media,
		MediaName:   "test_image.jpg",
		Width:       1000,
		Height:      800,
		Purpose:     models.PhotoThumbnail,
		ContentType: "image/jpeg",
		FileSize:    1024,
	}
	assert.NoError(t, db.Save(&mediaURL).Error)

	tempDir := t.TempDir()
	orig := utils.MediaCachePath()
	utils.ConfigureTestCache(tempDir)
	defer utils.ConfigureTestCache(orig)

	router := mux.NewRouter()
	RegisterPhotoRoutes(db, router)

	// -- Test cases --

	// Non-existent media_name => 404 (no auth required)
	t.Run("media not found", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/does_not_exist.jpg", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNotFound, rec.Code)
		assert.Equal(t, "404 - Media not found", rec.Body.String())
	})

	// Missing auth => 403
	t.Run("auth failure", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test_image.jpg", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusForbidden, rec.Code)
		assert.Contains(t, rec.Body.String(), "unauthorized")
	})

	// Cache miss + scan error => 500
	t.Run("scan failure yields 500", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test_image.jpg", nil)
		ctx := auth.AddUserToContext(req.Context(), user)
		req = req.WithContext(ctx)

		// ensure no cached file
		cachedPath, err := mediaURL.CachedPath()
		assert.NoError(t, err)
		os.Remove(cachedPath)

		// mock scan to fail
		origScan := scanner.ProcessSingleMediaFunc
		scanner.ProcessSingleMediaFunc = func(ctx context.Context, db *gorm.DB, m *models.Media) error {
			return fmt.Errorf("scan error")
		}
		defer func() { scanner.ProcessSingleMediaFunc = origScan }()

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Contains(t, rec.Body.String(), "server error")
	})

	// Cache hit => 200 with correct body and headers
	t.Run("cache hit serves file", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test_image.jpg", nil)
		ctx := auth.AddUserToContext(req.Context(), user)
		req = req.WithContext(ctx)

		// pre-create cached file
		cachedPath, err := mediaURL.CachedPath()
		assert.NoError(t, err)
		assert.NoError(t, os.MkdirAll(path.Dir(cachedPath), 0755))
		content := []byte("cached-binary")
		assert.NoError(t, os.WriteFile(cachedPath, content, 0644))

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "cached-binary", rec.Body.String())
		assert.Equal(t, "private, max-age=86400, immutable", rec.Header().Get("Cache-Control"))
	})
}
