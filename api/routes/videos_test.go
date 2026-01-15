package routes

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/graphql/models/actions"
	"github.com/photoview/photoview/api/test_utils"
	"github.com/photoview/photoview/api/utils"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// setTestCachePath temporarily sets a different media cache path for testing
// and returns a function to restore the original state
func setTestCachePath(tempPath string) func() {
	original := utils.GetTestCachePath()
	utils.ConfigureTestCache(tempPath)
	return func() {
		utils.ConfigureTestCache(original)
	}
}

// mockProcessSingleMedia replaces scanner.ProcessSingleMedia with a mock function during tests
// and returns a function to restore the original implementation
var originalProcessSingleMedia = processSingleMediaFn

func mockProcessSingleMedia(t *testing.T, shouldSucceed bool, mediaID int, albumID int) func() {
	// Save original implementation
	savedFn := processSingleMediaFn

	// Replace with mock implementation
	processSingleMediaFn = func(ctx context.Context, db *gorm.DB, fs afero.Fs, media *models.Media) error {
		// Check if context is already cancelled before starting work
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if shouldSucceed {
			// On success: create the expected video file in cache
			var mediaURLs []models.MediaURL
			if err := db.Where("media_id = ? AND purpose = ?", media.ID, models.VideoWeb).
				Find(&mediaURLs).Error; err != nil {
				return err
			}

			if len(mediaURLs) == 0 {
				return fmt.Errorf("no media URLs found")
			}

			// Get the cache path
			tempCachePath := utils.GetTestCachePath()
			albumDir := filepath.Join(tempCachePath, strconv.Itoa(albumID))
			mediaDir := filepath.Join(albumDir, strconv.Itoa(mediaID))
			if err := os.MkdirAll(mediaDir, 0755); err != nil {
				return err
			}

			videoPath := filepath.Join(mediaDir, mediaURLs[0].MediaName)
			if err := os.WriteFile(videoPath, []byte("mocked processed video content"), 0644); err != nil {
				return fmt.Errorf("failed to write mock video file: %w", err)
			}
			return nil
		}

		// On failure: return an error
		return fmt.Errorf("mock processing error")
	}

	// Return cleanup function
	return func() {
		processSingleMediaFn = savedFn
	}
}

func registerMockVideoRoutesForTesting(db *gorm.DB, fs afero.Fs, router *mux.Router, tempCachePath string) {
	router.HandleFunc("/{name}", func(w http.ResponseWriter, r *http.Request) {
		mediaName := mux.Vars(r)["name"]

		// Use no-op auth and test cache path
		handleVideoRequest(
			w, r, db, fs, mediaName,
			// Skip authentication for tests
			func(media *models.Media, db *gorm.DB, r *http.Request) (bool, string, int, error) {
				return true, "success", http.StatusOK, nil
			},
			// Use test cache path
			func(albumID, mediaID int, filename string) string {
				return path.Join(tempCachePath, strconv.Itoa(albumID), strconv.Itoa(mediaID), filename)
			},
		)
	})
}

// createTestResources creates all the necessary test resources for a single test case
// and returns cleanup functions to be called with t.Cleanup()
func createTestResources(t *testing.T, db *gorm.DB, testID string) (
	*models.User,
	*models.Album,
	*models.Media,
	*models.MediaURL,
	string, // mediaName
	string, // cachePath
	string, // shareToken
	string, // tokenPassword
) {
	// Create test user with unique username
	user := &models.User{
		Username: fmt.Sprintf("testuser-%s", testID),
	}
	require.NoError(t, db.Create(user).Error)
	t.Cleanup(func() {
		db.Unscoped().Delete(user)
	})

	// Create test album with unique title and path
	album := &models.Album{
		Title: fmt.Sprintf("Test Album %s", testID),
		Path:  fmt.Sprintf("/test/album/path/%s", testID),
	}
	require.NoError(t, db.Create(album).Error)
	t.Cleanup(func() {
		db.Unscoped().Delete(album)
	})

	// Establish ownership via many-to-many relationship
	require.NoError(t, db.Model(album).Association("Owners").Append(user))
	t.Cleanup(func() {
		db.Model(album).Association("Owners").Clear()
	})

	// Create unique media name for this test
	mediaName := fmt.Sprintf("video-%s.mp4", testID)

	// Create media with VideoWeb purpose
	media := &models.Media{
		Title:    fmt.Sprintf("Test Video %s", testID),
		Path:     filepath.Join(t.TempDir(), mediaName),
		PathHash: fmt.Sprintf("testhash-%s", testID),
		AlbumID:  album.ID,
		Album:    *album,
		DateShot: time.Now(),
		Type:     "video",
	}
	require.NoError(t, db.Create(media).Error)
	t.Cleanup(func() {
		db.Unscoped().Delete(media)
	})

	// Create media URL entry
	mediaURL := &models.MediaURL{
		MediaID:     media.ID,
		Media:       media,
		MediaName:   mediaName,
		Width:       1920,
		Height:      1080,
		Purpose:     models.VideoWeb,
		ContentType: "video/mp4",
		FileSize:    1024,
	}
	require.NoError(t, db.Create(mediaURL).Error)
	t.Cleanup(func() {
		db.Unscoped().Delete(mediaURL)
	})

	// Create a unique cache path for this test
	cachePath := filepath.Join(t.TempDir(), fmt.Sprintf("cache-%s", testID))
	require.NoError(t, os.MkdirAll(cachePath, 0755))
	t.Cleanup(func() {
		os.RemoveAll(cachePath)
	})

	// Prepare share token for auth tests
	tokenPassword := fmt.Sprintf("secret-password-%s", testID)
	expiry := time.Now().Add(24 * time.Hour)
	shareToken, err := actions.AddMediaShare(db, user, media.ID, &expiry, &tokenPassword)
	require.NoError(t, err)
	t.Cleanup(func() {
		db.Unscoped().Delete(shareToken)
	})

	return user, album, media, mediaURL, mediaName, cachePath, shareToken.Value, tokenPassword
}

func TestVideoRoutes(t *testing.T) {
	// Ensure original function is always restored
	defer func() {
		processSingleMediaFn = originalProcessSingleMedia
	}()

	// Setup test database
	db := test_utils.DatabaseTest(t)
	fs := test_utils.FilesystemTest(t)

	// Define test cases
	testCases := []struct {
		name     string
		testFunc func(*testing.T, *gorm.DB)
	}{
		{
			name: "Valid video retrieval",
			testFunc: func(t *testing.T, db *gorm.DB) {
				// Create unique resources for this test
				_, album, media, _, mediaName, cachePath, _, _ := createTestResources(t, db, "valid")

				// Setup cache path for this test
				restorePath := setTestCachePath(cachePath)
				t.Cleanup(restorePath)

				// Create cache directory and file
				albumDir := filepath.Join(cachePath, strconv.Itoa(int(album.ID)))
				mediaDir := filepath.Join(albumDir, strconv.Itoa(int(media.ID)))
				require.NoError(t, os.MkdirAll(mediaDir, 0755))

				videoPath := filepath.Join(mediaDir, mediaName)
				require.NoError(t, os.WriteFile(videoPath, []byte("test video content"), 0644))

				// Create mock router without auth for this test
				router := mux.NewRouter()
				registerMockVideoRoutesForTesting(db, fs, router, cachePath)

				// Make request
				req := httptest.NewRequest("GET", "/"+mediaName, nil)
				rr := httptest.NewRecorder()
				router.ServeHTTP(rr, req)

				// Validate response
				assert.Equal(t, http.StatusOK, rr.Code)
				assert.Equal(t, "test video content", rr.Body.String())
			},
		},
		{
			name: "Video not found",
			testFunc: func(t *testing.T, db *gorm.DB) {
				// Create unique resources for this test
				_, _, _, _, _, cachePath, _, _ := createTestResources(t, db, "notfound")

				// Setup cache path for this test
				restorePath := setTestCachePath(cachePath)
				t.Cleanup(restorePath)

				// Create mock router without auth for this test
				router := mux.NewRouter()
				registerMockVideoRoutesForTesting(db, fs, router, cachePath)

				// Make request with nonexistent video name
				req := httptest.NewRequest("GET", "/nonexistent.mp4", nil)
				rr := httptest.NewRecorder()
				router.ServeHTTP(rr, req)

				// Validate response
				assert.Equal(t, http.StatusNotFound, rr.Code)
				assert.Equal(t, "not found", rr.Body.String())
			},
		},
		{
			name: "Authentication with share token",
			testFunc: func(t *testing.T, db *gorm.DB) {
				// Create unique resources for this test
				_, album, media, _, mediaName, cachePath, tokenValue, tokenPassword := createTestResources(t, db, "auth")

				// Setup cache path for this test
				restorePath := setTestCachePath(cachePath)
				t.Cleanup(restorePath)

				// Create the file in cache
				albumDir := filepath.Join(cachePath, strconv.Itoa(int(album.ID)))
				mediaDir := filepath.Join(albumDir, strconv.Itoa(int(media.ID)))
				require.NoError(t, os.MkdirAll(mediaDir, 0755))

				videoPath := filepath.Join(mediaDir, mediaName)
				require.NoError(t, os.WriteFile(videoPath, []byte("test video content"), 0644))

				// Create real router with auth for this test
				router := mux.NewRouter()
				RegisterVideoRoutes(db, fs, router)

				// Make request with token
				req := httptest.NewRequest("GET", "/"+mediaName+"?token="+tokenValue, nil)
				cookie := http.Cookie{
					Name:  fmt.Sprintf("share-token-pw-%s", tokenValue),
					Value: tokenPassword,
				}
				req.AddCookie(&cookie)

				rr := httptest.NewRecorder()
				router.ServeHTTP(rr, req)

				// Validate response
				assert.Equal(t, http.StatusOK, rr.Code)
				assert.Equal(t, "test video content", rr.Body.String())
			},
		},
		{
			name: "Multiple media URLs with same name",
			testFunc: func(t *testing.T, db *gorm.DB) {
				// Create unique resources for this test
				_, album, media, _, mediaName, cachePath, _, _ := createTestResources(t, db, "multiple")

				// Setup cache path for this test
				restorePath := setTestCachePath(cachePath)
				t.Cleanup(restorePath)

				// Create second mediaURL with same name
				mediaURL2 := &models.MediaURL{
					MediaID:     media.ID,
					Media:       media,
					MediaName:   mediaName, // Same name
					Width:       1280,
					Height:      720,
					Purpose:     models.VideoWeb,
					ContentType: "video/mp4",
					FileSize:    512,
				}
				require.NoError(t, db.Create(mediaURL2).Error)
				t.Cleanup(func() {
					db.Unscoped().Delete(mediaURL2)
				})

				// Create cache directory and file
				albumDir := filepath.Join(cachePath, strconv.Itoa(int(album.ID)))
				mediaDir := filepath.Join(albumDir, strconv.Itoa(int(media.ID)))
				require.NoError(t, os.MkdirAll(mediaDir, 0755))

				videoPath := filepath.Join(mediaDir, mediaName)
				require.NoError(t, os.WriteFile(videoPath, []byte("test video content"), 0644))

				// Create mock router without auth for this test
				router := mux.NewRouter()
				registerMockVideoRoutesForTesting(db, fs, router, cachePath)

				// Make request
				req := httptest.NewRequest("GET", "/"+mediaName, nil)
				rr := httptest.NewRecorder()
				router.ServeHTTP(rr, req)

				// Validate response
				assert.Equal(t, http.StatusOK, rr.Code)
				assert.Equal(t, "test video content", rr.Body.String())
			},
		},
		{
			name: "Video file not in cache, processing succeeds",
			testFunc: func(t *testing.T, db *gorm.DB) {
				// Create unique resources for this test
				_, album, media, _, mediaName, cachePath, _, _ := createTestResources(t, db, "process-success")

				// Setup cache path for this test
				restorePath := setTestCachePath(cachePath)
				t.Cleanup(restorePath)

				// Ensure cache directory exists but file doesn't exist
				albumDir := filepath.Join(cachePath, strconv.Itoa(int(album.ID)))
				mediaDir := filepath.Join(albumDir, strconv.Itoa(int(media.ID)))
				require.NoError(t, os.MkdirAll(mediaDir, 0755))

				// Mock processing to succeed
				restoreProcessingFn := mockProcessSingleMedia(t, true, int(media.ID), int(album.ID))
				t.Cleanup(restoreProcessingFn)

				// Create mock router without auth for this test
				router := mux.NewRouter()
				registerMockVideoRoutesForTesting(db, fs, router, cachePath)

				// Make request
				req := httptest.NewRequest("GET", "/"+mediaName, nil)
				rr := httptest.NewRecorder()
				router.ServeHTTP(rr, req)

				// Validate response
				assert.Equal(t, http.StatusOK, rr.Code)
				assert.Equal(t, "mocked processed video content", rr.Body.String())
			},
		},
		{
			name: "Video file not in cache, processing fails",
			testFunc: func(t *testing.T, db *gorm.DB) {
				// Create unique resources for this test
				_, album, media, _, mediaName, cachePath, _, _ := createTestResources(t, db, "process-fail")

				// Setup cache path for this test
				restorePath := setTestCachePath(cachePath)
				t.Cleanup(restorePath)

				// Ensure cache directory exists but file doesn't exist
				albumDir := filepath.Join(cachePath, strconv.Itoa(int(album.ID)))
				mediaDir := filepath.Join(albumDir, strconv.Itoa(int(media.ID)))
				require.NoError(t, os.MkdirAll(mediaDir, 0755))

				// Mock processing to fail
				restoreProcessingFn := mockProcessSingleMedia(t, false, int(media.ID), int(album.ID))
				t.Cleanup(restoreProcessingFn)

				// Create mock router without auth for this test
				router := mux.NewRouter()
				registerMockVideoRoutesForTesting(db, fs, router, cachePath)

				// Make request
				req := httptest.NewRequest("GET", "/"+mediaName, nil)
				rr := httptest.NewRecorder()
				router.ServeHTTP(rr, req)

				// Validate response
				assert.Equal(t, http.StatusInternalServerError, rr.Code)
			},
		},
		{
			name: "Context cancellation during processing",
			testFunc: func(t *testing.T, db *gorm.DB) {
				_, album, media, _, mediaName, cachePath, _, _ := createTestResources(t, db, "cancellation")

				// Setup cache path for this test
				restorePath := setTestCachePath(cachePath)
				t.Cleanup(restorePath)

				// Ensure cache directory exists but file doesn't exist to trigger processing
				albumDir := filepath.Join(cachePath, strconv.Itoa(int(album.ID)))
				mediaDir := filepath.Join(albumDir, strconv.Itoa(int(media.ID)))
				require.NoError(t, os.MkdirAll(mediaDir, 0755))

				// Create cancellable context
				ctx, cancel := context.WithCancel(context.Background())

				// Mock processing that simulates context cancellation
				savedFn := processSingleMediaFn
				processSingleMediaFn = func(reqCtx context.Context, db *gorm.DB, fs afero.Fs, media *models.Media) error {
					cancel()
					return fmt.Errorf("processing interrupted by cancellation")
				}
				t.Cleanup(func() { processSingleMediaFn = savedFn })

				// Create request with cancelled context
				req := httptest.NewRequest("GET", "/video/"+mediaName, nil)
				req = req.WithContext(ctx)
				w := httptest.NewRecorder()

				// Use testing router without auth
				mockRouter := mux.NewRouter().PathPrefix("/video").Subrouter()
				registerMockVideoRoutesForTesting(db, fs, mockRouter, cachePath)

				mockRouter.ServeHTTP(w, req)

				// When context is cancelled, processing should be cancelled
				// 1. Status remains default 200 (no explicit status written)
				assert.Equal(t, http.StatusOK, w.Code, "Status should remain default when context cancelled")

				// 2. Response body should be empty (no video content served)
				assert.Empty(t, w.Body.String(), "Response body should be empty when context cancelled")
			},
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.testFunc(t, db)
		})
	}
}
