package routes

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strconv"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/graphql/models/actions"
	"github.com/photoview/photoview/api/test_utils"
	"github.com/photoview/photoview/api/utils"
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

func mockProcessSingleMedia(t *testing.T, shouldSucceed bool) func() {
	// Replace with mock implementation
	processSingleMediaFn = func(db *gorm.DB, media *models.Media) error {
		if shouldSucceed {
			// On success: create the expected video file in cache
			mediaURLs := []models.MediaURL{}
			if err := db.Where("media_id = ? AND purpose = ?", media.ID, models.VideoWeb).Find(&mediaURLs).Error; err != nil {
				return err
			}

			if len(mediaURLs) == 0 {
				return fmt.Errorf("no media URLs found")
			}

			// Get the cache path
			tempCachePath := utils.GetTestCachePath()
			albumDir := path.Join(tempCachePath, strconv.Itoa(int(media.AlbumID)))
			mediaDir := path.Join(albumDir, strconv.Itoa(int(mediaURLs[0].MediaID)))
			if err := os.MkdirAll(mediaDir, 0755); err != nil {
				return err
			}

			videoPath := path.Join(mediaDir, mediaURLs[0].MediaName)
			return os.WriteFile(videoPath, []byte("mocked processed video content"), 0644)
		}

		// On failure: return an error
		return fmt.Errorf("mock processing error")
	}

	// Return cleanup function
	return func() {
		processSingleMediaFn = originalProcessSingleMedia
	}
}

func TestVideoRoutes(t *testing.T) {
	defer func() {
		processSingleMediaFn = originalProcessSingleMedia
	}()

	// Setup test database
	db := test_utils.DatabaseTest(t)

	// Setup test cache directory
	tempCachePath := t.TempDir()
	restorePath := setTestCachePath(tempCachePath)
	defer restorePath()

	// Create test user
	user := &models.User{
		Username: "testuser",
	}
	require.NoError(t, db.Create(user).Error)

	// Create test album
	album := &models.Album{
		Title: "Test Album",
		Path:  "/test/album/path",
	}
	require.NoError(t, db.Create(album).Error)

	// Establish ownership via many-to-many relationship
	require.NoError(t, db.Model(album).Association("Owners").Append(user))

	// Create media with VideoWeb purpose
	media := &models.Media{
		Title:    "Test Video",
		Path:     path.Join(t.TempDir(), "video.mp4"),
		PathHash: "testhash1",
		AlbumID:  album.ID,
		Album:    *album,
		DateShot: time.Now(),
		Type:     "video",
	}
	require.NoError(t, db.Create(media).Error)

	// Create media URL entry
	mediaURL := &models.MediaURL{
		MediaID:     media.ID,
		Media:       media,
		MediaName:   "video.mp4",
		Width:       1920,
		Height:      1080,
		Purpose:     models.VideoWeb,
		ContentType: "video/mp4",
		FileSize:    1024,
	}
	require.NoError(t, db.Create(mediaURL).Error)

	// Create another media with no URLs
	mediaNoURLs := &models.Media{
		Title:    "Video Without URLs",
		Path:     path.Join(t.TempDir(), "no-urls.mp4"),
		PathHash: "testhash2",
		AlbumID:  album.ID,
		DateShot: time.Now(),
		Type:     "video",
	}
	require.NoError(t, db.Create(mediaNoURLs).Error)

	// Prepare share token for auth tests
	tokenPassword := "secret-password"
	expiry := time.Now().Add(24 * time.Hour)
	shareToken, err := actions.AddMediaShare(db, user, media.ID, &expiry, &tokenPassword)
	require.NoError(t, err)

	// Create two routers - one with real handler for auth tests, one with mock handler for other tests
	realRouter := mux.NewRouter()
	RegisterVideoRoutes(db, realRouter)

	mockRouter := mux.NewRouter()
	registerMockVideoRoutesForTesting(db, mockRouter, tempCachePath)

	// Create test cases
	testCases := []struct {
		name         string
		url          string
		useRealAuth  bool // To determine which router to use
		setupFunc    func(t *testing.T) *httptest.ResponseRecorder
		validateFunc func(t *testing.T, rr *httptest.ResponseRecorder)
	}{
		{
			name:        "Valid video retrieval",
			url:         "/video.mp4",
			useRealAuth: false, // Use mock router for non-auth tests
			setupFunc: func(t *testing.T) *httptest.ResponseRecorder {
				// Create cache directory and file
				albumDir := path.Join(tempCachePath, strconv.Itoa(int(album.ID)))
				mediaDir := path.Join(albumDir, strconv.Itoa(int(mediaURL.MediaID)))
				require.NoError(t, os.MkdirAll(mediaDir, 0755))

				videoPath := path.Join(mediaDir, mediaURL.MediaName)
				require.NoError(t, os.WriteFile(videoPath, []byte("test video content"), 0644))

				return httptest.NewRecorder()
			},
			validateFunc: func(t *testing.T, rr *httptest.ResponseRecorder) {
				albumDir := path.Join(tempCachePath, strconv.Itoa(int(album.ID)))
				filePath := path.Join(albumDir, strconv.Itoa(int(mediaURL.MediaID)), mediaURL.MediaName)
				_, err := os.Stat(filePath)
				require.NoError(t, err, "File does not exist: %s", filePath)

				assert.Equal(t, http.StatusOK, rr.Code)
				assert.Equal(t, "test video content", rr.Body.String())

				t.Cleanup(func() {
					require.NoError(t, os.RemoveAll(albumDir))
				})
			},
		},
		{
			name:        "Video not found",
			url:         "/nonexistent.mp4",
			useRealAuth: false, // Use mock router for non-auth tests
			setupFunc: func(t *testing.T) *httptest.ResponseRecorder {
				return httptest.NewRecorder()
			},
			validateFunc: func(t *testing.T, rr *httptest.ResponseRecorder) {
				filePath := path.Join(tempCachePath, strconv.Itoa(int(album.ID)), strconv.Itoa(int(mediaURL.MediaID)), mediaURL.MediaName)
				_, err := os.Stat(filePath)
				require.Error(t, err, "File exists: %s", filePath)

				assert.Equal(t, http.StatusNotFound, rr.Code)
				assert.Equal(t, "not found", rr.Body.String())
			},
		},
		{
			name:        "Filesystem permission error",
			url:         "/video.mp4",
			useRealAuth: false,
			setupFunc: func(t *testing.T) *httptest.ResponseRecorder {
				// Prepare cache dir and file
				albumDir := path.Join(tempCachePath, strconv.Itoa(int(album.ID)))
				mediaDir := path.Join(albumDir, strconv.Itoa(int(mediaURL.MediaID)))
				require.NoError(t, os.MkdirAll(mediaDir, 0755))
				videoPath := path.Join(mediaDir, mediaURL.MediaName)
				// Create a directory instead of a file - this will cause the file to exist
				// (passing the os.Stat check) but will fail when http.ServeFile tries to serve it
				require.NoError(t, os.Mkdir(videoPath, 0755))

				// Verify our setup works correctly
				fi, err := os.Stat(videoPath)
				require.NoError(t, err, "Directory should exist")
				require.True(t, fi.IsDir(), "Path should be a directory, not a file")

				return httptest.NewRecorder()
			},
			validateFunc: func(t *testing.T, rr *httptest.ResponseRecorder) {
				albumDir := path.Join(tempCachePath, strconv.Itoa(int(album.ID)))
				folderPath := path.Join(albumDir, strconv.Itoa(int(mediaURL.MediaID)), mediaURL.MediaName)
				files, err := os.ReadDir(folderPath)
				require.NoError(t, err, "Failed to read folder: %s", folderPath)
				require.Empty(t, files, "Folder is not empty: %s", folderPath)

				assert.Equal(t, http.StatusInternalServerError, rr.Code)

				t.Cleanup(func() {
					require.NoError(t, os.RemoveAll(albumDir))
				})
			},
		},
		{
			name:        "Authentication with share token",
			url:         fmt.Sprintf("/video.mp4?token=%s", shareToken.Value),
			useRealAuth: true, // Use real router for auth tests
			setupFunc: func(t *testing.T) *httptest.ResponseRecorder {
				// Create the file in cache
				albumDir := path.Join(tempCachePath, strconv.Itoa(int(album.ID)))
				mediaDir := path.Join(albumDir, strconv.Itoa(int(mediaURL.MediaID)))
				require.NoError(t, os.MkdirAll(mediaDir, 0755))

				videoPath := path.Join(mediaDir, mediaURL.MediaName)
				require.NoError(t, os.WriteFile(videoPath, []byte("test video content"), 0644))

				return httptest.NewRecorder()
			},
			validateFunc: func(t *testing.T, rr *httptest.ResponseRecorder) {
				albumDir := path.Join(tempCachePath, strconv.Itoa(int(album.ID)))
				filePath := path.Join(albumDir, strconv.Itoa(int(mediaURL.MediaID)), mediaURL.MediaName)
				_, err := os.Stat(filePath)
				require.NoError(t, err, "File does not exist: %s", filePath)

				assert.Equal(t, http.StatusOK, rr.Code)
				assert.Equal(t, "test video content", rr.Body.String())

				t.Cleanup(func() {
					require.NoError(t, os.RemoveAll(albumDir))
				})
			},
		},
		{
			name:        "Multiple media URLs with same name",
			url:         "/video.mp4",
			useRealAuth: false, // Use mock router for non-auth tests
			setupFunc: func(t *testing.T) *httptest.ResponseRecorder {
				// Create second mediaURL with same name
				mediaURL2 := &models.MediaURL{
					MediaID:     media.ID,
					Media:       media,
					MediaName:   "video.mp4", // Same name
					Width:       1280,
					Height:      720,
					Purpose:     models.VideoWeb,
					ContentType: "video/mp4",
					FileSize:    512,
				}
				require.NoError(t, db.Create(mediaURL2).Error)

				// Create cache directory and file
				albumDir := path.Join(tempCachePath, strconv.Itoa(int(album.ID)))
				mediaDir := path.Join(albumDir, strconv.Itoa(int(mediaURL.MediaID)))
				require.NoError(t, os.MkdirAll(mediaDir, 0755))

				videoPath := path.Join(mediaDir, mediaURL.MediaName)
				require.NoError(t, os.WriteFile(videoPath, []byte("test video content"), 0644))

				return httptest.NewRecorder()
			},
			validateFunc: func(t *testing.T, rr *httptest.ResponseRecorder) {
				albumDir := path.Join(tempCachePath, strconv.Itoa(int(album.ID)))
				filePath := path.Join(albumDir, strconv.Itoa(int(mediaURL.MediaID)), mediaURL.MediaName)
				_, err := os.Stat(filePath)
				require.NoError(t, err, "File does not exist: %s", filePath)

				assert.Equal(t, http.StatusOK, rr.Code)
				assert.Equal(t, "test video content", rr.Body.String())

				t.Cleanup(func() {
					require.NoError(t, os.RemoveAll(albumDir))
					db.Unscoped().Where("media_name = ? AND id != ?", "video.mp4", mediaURL.ID).Delete(&models.MediaURL{})
				})
			},
		},
		{
			name:        "Video file not in cache, processing succeeds",
			url:         "/video.mp4",
			useRealAuth: false,
			setupFunc: func(t *testing.T) *httptest.ResponseRecorder {
				// Ensure cache directory exists but file doesn't exist
				albumDir := path.Join(tempCachePath, strconv.Itoa(int(album.ID)))
				mediaDir := path.Join(albumDir, strconv.Itoa(int(mediaURL.MediaID)))
				require.NoError(t, os.MkdirAll(mediaDir, 0755))

				t.Cleanup(func() {
					mockProcessSingleMedia(t, true)
				})

				return httptest.NewRecorder()
			},
			validateFunc: func(t *testing.T, rr *httptest.ResponseRecorder) {
				albumDir := path.Join(tempCachePath, strconv.Itoa(int(album.ID)))
				filePath := path.Join(albumDir, strconv.Itoa(int(mediaURL.MediaID)), mediaURL.MediaName)
				_, err := os.Stat(filePath)
				require.NoError(t, err, "File does not exist: %s", filePath)

				assert.Equal(t, http.StatusOK, rr.Code)
				assert.Equal(t, "mocked processed video content", rr.Body.String())

				t.Cleanup(func() {
					require.NoError(t, os.RemoveAll(albumDir))
				})
			},
		},
		{
			name:        "Video file not in cache, processing fails",
			url:         "/video.mp4",
			useRealAuth: false,
			setupFunc: func(t *testing.T) *httptest.ResponseRecorder {
				// Ensure cache directory exists but file doesn't exist
				albumDir := path.Join(tempCachePath, strconv.Itoa(int(album.ID)))
				mediaDir := path.Join(albumDir, strconv.Itoa(int(mediaURL.MediaID)))
				require.NoError(t, os.MkdirAll(mediaDir, 0755))

				t.Cleanup(func() {
					mockProcessSingleMedia(t, false)
				})

				return httptest.NewRecorder()
			},
			validateFunc: func(t *testing.T, rr *httptest.ResponseRecorder) {
				albumDir := path.Join(tempCachePath, strconv.Itoa(int(album.ID)))
				mediaDir := path.Join(albumDir, strconv.Itoa(int(mediaURL.MediaID)))
				files, err := os.ReadDir(mediaDir)
				require.NoError(t, err, "Failed to read folder: %s", mediaDir)
				require.Empty(t, files, "Folder is not empty: %s", mediaDir, files)

				assert.Equal(t, http.StatusInternalServerError, rr.Code)

				t.Cleanup(func() {
					require.NoError(t, os.RemoveAll(albumDir))
				})
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rr := tc.setupFunc(t)
			req := httptest.NewRequest("GET", tc.url, nil)

			// For auth tests, add proper cookie
			if tc.useRealAuth {
				cookie := http.Cookie{
					Name:  fmt.Sprintf("share-token-pw-%s", shareToken.Value),
					Value: tokenPassword,
				}
				req.AddCookie(&cookie)

				// Use real router for auth test
				realRouter.ServeHTTP(rr, req)
			} else {
				// Use mock router for non-auth tests
				mockRouter.ServeHTTP(rr, req)
			}

			tc.validateFunc(t, rr)
		})
	}
}
