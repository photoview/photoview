package routes

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"image/jpeg"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/gorilla/mux"
	"github.com/photoview/photoview/api/graphql/auth"
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/test_utils"
	"github.com/stretchr/testify/assert"
)

func validJPEGBytes(t *testing.T) []byte {
	t.Helper()

	img := image.NewNRGBA(image.Rect(0, 0, 4, 4))
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			img.Set(x, y, color.NRGBA{R: 255, A: 255})
		}
	}

	var buf bytes.Buffer
	assert.NoError(t, jpeg.Encode(&buf, img, nil))
	return buf.Bytes()
}

// buildUploadRequest builds a multipart request to POST /{albumId}, with
// each entry in files keyed by the relative path used as the form field
// name (matching the client-side contract).
func buildUploadRequest(t *testing.T, albumID int, files map[string][]byte) *http.Request {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for relPath, content := range files {
		part, err := writer.CreateFormFile(relPath, filepath.Base(relPath))
		assert.NoError(t, err)
		_, err = part.Write(content)
		assert.NoError(t, err)
	}
	assert.NoError(t, writer.Close())

	req := httptest.NewRequest("POST", "/"+strconv.Itoa(albumID), &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

func TestUploadRoute(t *testing.T) {
	db := test_utils.DatabaseTest(t)

	// Stub out the real scanner queue: it starts a background worker
	// goroutine on first use and isn't needed to verify this route's own
	// permission/validation/filesystem behavior.
	origAddAlbumToQueue := addAlbumToQueue
	addAlbumToQueue = func(album *models.Album) error { return nil }
	t.Cleanup(func() { addAlbumToQueue = origAddAlbumToQueue })

	uploader, err := models.RegisterUser(db, "uploader", nil, false)
	assert.NoError(t, err)

	nonUploader, err := models.RegisterUser(db, "non_uploader", nil, false)
	assert.NoError(t, err)

	albumPath := t.TempDir()
	album := models.Album{Title: "album", Path: albumPath}
	assert.NoError(t, db.Save(&album).Error)
	assert.NoError(t, db.Create(&models.UserAlbums{
		UserID: uploader.ID, AlbumID: album.ID, Level: models.AlbumPermissionLevelUpload,
	}).Error)
	assert.NoError(t, db.Create(&models.UserAlbums{
		UserID: nonUploader.ID, AlbumID: album.ID, Level: models.AlbumPermissionLevelRead,
	}).Error)

	router := mux.NewRouter()
	RegisterUploadRoutes(db, router)

	t.Run("unauthenticated request is rejected", func(t *testing.T) {
		req := buildUploadRequest(t, album.ID, map[string][]byte{"photo.jpg": validJPEGBytes(t)})
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusForbidden, rec.Code)
	})

	t.Run("user with only Read level is rejected", func(t *testing.T) {
		req := buildUploadRequest(t, album.ID, map[string][]byte{"photo.jpg": validJPEGBytes(t)})
		req = req.WithContext(auth.AddUserToContext(req.Context(), nonUploader))
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusForbidden, rec.Code)
	})

	t.Run("valid image is stored and reported ok", func(t *testing.T) {
		req := buildUploadRequest(t, album.ID, map[string][]byte{"photo.jpg": validJPEGBytes(t)})
		req = req.WithContext(auth.AddUserToContext(req.Context(), uploader))
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var resp uploadResponse
		assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		if assert.Len(t, resp.Results, 1) {
			assert.Equal(t, "ok", resp.Results[0].Status)
		}

		_, statErr := os.Stat(filepath.Join(albumPath, "photo.jpg"))
		assert.NoError(t, statErr)
	})

	t.Run("uploading over an existing file is rejected, not overwritten", func(t *testing.T) {
		existingPath := filepath.Join(albumPath, "existing.jpg")
		assert.NoError(t, os.WriteFile(existingPath, []byte("original content"), 0o644))

		req := buildUploadRequest(t, album.ID, map[string][]byte{"existing.jpg": validJPEGBytes(t)})
		req = req.WithContext(auth.AddUserToContext(req.Context(), uploader))
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var resp uploadResponse
		assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		if assert.Len(t, resp.Results, 1) {
			assert.Equal(t, "rejected", resp.Results[0].Status)
		}

		onDisk, err := os.ReadFile(existingPath)
		assert.NoError(t, err)
		assert.Equal(t, "original content", string(onDisk), "existing file must not be overwritten")
	})

	t.Run("whole-folder upload creates intermediate directories", func(t *testing.T) {
		req := buildUploadRequest(t, album.ID, map[string][]byte{
			"Vacation/Day1/beach.jpg": validJPEGBytes(t),
		})
		req = req.WithContext(auth.AddUserToContext(req.Context(), uploader))
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		_, statErr := os.Stat(filepath.Join(albumPath, "Vacation", "Day1", "beach.jpg"))
		assert.NoError(t, statErr)
	})

	t.Run("unsupported file type is rejected", func(t *testing.T) {
		req := buildUploadRequest(t, album.ID, map[string][]byte{"notes.txt": []byte("hello")})
		req = req.WithContext(auth.AddUserToContext(req.Context(), uploader))
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var resp uploadResponse
		assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		if assert.Len(t, resp.Results, 1) {
			assert.Equal(t, "rejected", resp.Results[0].Status)
		}

		_, statErr := os.Stat(filepath.Join(albumPath, "notes.txt"))
		assert.True(t, os.IsNotExist(statErr))
	})

	t.Run("path traversal field name is rejected and nothing escapes the album dir", func(t *testing.T) {
		req := buildUploadRequest(t, album.ID, map[string][]byte{"../../evil.jpg": validJPEGBytes(t)})
		req = req.WithContext(auth.AddUserToContext(req.Context(), uploader))
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var resp uploadResponse
		assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		if assert.Len(t, resp.Results, 1) {
			assert.Equal(t, "rejected", resp.Results[0].Status)
		}

		_, statErr := os.Stat(filepath.Join(filepath.Dir(filepath.Dir(albumPath)), "evil.jpg"))
		assert.True(t, os.IsNotExist(statErr))
	})

	t.Run("unknown album id is forbidden, same as a permission denial", func(t *testing.T) {
		req := buildUploadRequest(t, 999999, map[string][]byte{"photo.jpg": validJPEGBytes(t)})
		req = req.WithContext(auth.AddUserToContext(req.Context(), uploader))
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusForbidden, rec.Code)
	})
}
