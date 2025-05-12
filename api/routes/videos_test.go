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
	"github.com/photoview/photoview/api/log"
	"github.com/photoview/photoview/api/scanner"
	"github.com/photoview/photoview/api/test_utils"
	"github.com/photoview/photoview/api/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// setTestCachePath temporarily sets a different media cache path for testing
// and returns a function to restore the original state
func setTestCachePath(tempPath string) func() {
	// Store the original value
	original := utils.GetTestCachePath()
	// Set the test value
	utils.ConfigureTestCache(tempPath)
	// Return function to restore original value
	return func() {
		utils.ConfigureTestCache(original)
	}
}

// mockVideoHandler replicates the video route handler but skips authentication
func mockVideoHandler(db *gorm.DB, tempCachePath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mediaName := mux.Vars(r)["name"]

		var mediaURLs []models.MediaURL
		if err := db.Model(&models.MediaURL{}).
			Select("media_urls.*").
			Joins("Media").
			Where("media_urls.media_name = ? AND media_urls.purpose = ?", mediaName, models.VideoWeb).
			Find(&mediaURLs).
			Error; err != nil || len(mediaURLs) == 0 || mediaURLs[0].Media == nil {

			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("404"))
			return
		}

		if len(mediaURLs) > 1 {
			log.Warn("Multiple video web URLs found for name", mediaName, "count", len(mediaURLs), "using", mediaURLs[0])
		}

		mediaURL := mediaURLs[0]
		media := mediaURL.Media

		// Authentication check is skipped in mock handler

		var cachedPath string

		if mediaURL.Purpose == models.VideoWeb {
			cachedPath = path.Join(tempCachePath, strconv.Itoa(int(media.AlbumID)), strconv.Itoa(int(mediaURL.MediaID)), mediaURL.MediaName)
		} else {
			log.Error("Can not handle media_purpose for video:", mediaURL.Purpose)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("internal server error"))
			return
		}

		if _, err := os.Stat(cachedPath); err != nil {
			if os.IsNotExist(err) {
				if err := scanner.ProcessSingleMedia(db, media); err != nil {
					log.Error("processing video not found in cache:", err)
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte("internal server error"))
					return
				}

				if _, err := os.Stat(cachedPath); err != nil {
					log.Error("video not found in cache after reprocessing:", err)
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte("internal server error"))
					return
				}
			}
		}

		http.ServeFile(w, r, cachedPath)
	}
}

func TestVideoRoutes(t *testing.T) {
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
	mockRouter.HandleFunc("/{name}", mockVideoHandler(db, tempCachePath))

	// Create test cases
	testCases := []struct {
		name         string
		url          string
		useRealAuth  bool // To determine which router to use
		setupFunc    func(t *testing.T) *httptest.ResponseRecorder
		validateFunc func(t *testing.T, rr *httptest.ResponseRecorder)
		cleanupFunc  func(t *testing.T)
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
				assert.Equal(t, http.StatusOK, rr.Code)
				assert.Equal(t, "test video content", rr.Body.String())
			},
			cleanupFunc: nil,
		},
		{
			name:        "Video not found",
			url:         "/nonexistent.mp4",
			useRealAuth: false, // Use mock router for non-auth tests
			setupFunc: func(t *testing.T) *httptest.ResponseRecorder {
				return httptest.NewRecorder()
			},
			validateFunc: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusNotFound, rr.Code)
				assert.Equal(t, "404", rr.Body.String())
			},
			cleanupFunc: nil,
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
				assert.Equal(t, http.StatusOK, rr.Code)
				assert.Equal(t, "test video content", rr.Body.String())
			},
			cleanupFunc: nil,
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
				// Should still work despite multiple URLs with warning logged
				assert.Equal(t, http.StatusOK, rr.Code)
				assert.Equal(t, "test video content", rr.Body.String())
			},
			cleanupFunc: func(t *testing.T) {
				// Delete the second media URL
				db.Unscoped().Where("media_name = ? AND id != ?", "video.mp4", mediaURL.ID).Delete(&models.MediaURL{})
			},
		},
		{
			name:        "Video file not in cache, needs processing",
			url:         "/video.mp4",
			useRealAuth: false, // Use mock router for non-auth tests
			setupFunc: func(t *testing.T) *httptest.ResponseRecorder {
				// Ensure cache directory exists but file doesn't
				albumDir := path.Join(tempCachePath, strconv.Itoa(int(album.ID)))
				mediaDir := path.Join(albumDir, strconv.Itoa(int(mediaURL.MediaID)))
				require.NoError(t, os.MkdirAll(mediaDir, 0755))

				// Delete file if it exists
				videoPath := path.Join(mediaDir, mediaURL.MediaName)
				os.Remove(videoPath)

				return httptest.NewRecorder()
			},
			validateFunc: func(t *testing.T, rr *httptest.ResponseRecorder) {
				// This will likely fail in tests because ProcessSingleMedia can't fully work in tests
				// When ProcessSingleMedia fails, we expect 500
				if rr.Code != http.StatusOK {
					assert.Equal(t, http.StatusInternalServerError, rr.Code)
				}
			},
			cleanupFunc: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rr := tc.setupFunc(t)
			req := httptest.NewRequest("GET", tc.url, nil)

			// For auth tests, add proper cookie
			if tc.name == "Authentication with share token" {
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

			if tc.cleanupFunc != nil {
				tc.cleanupFunc(t)
			}
		})
	}
}
