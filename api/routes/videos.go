package routes

import (
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"unicode"

	"github.com/gorilla/mux"
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/log"
	"github.com/photoview/photoview/api/scanner"
	"github.com/photoview/photoview/api/utils"
	"gorm.io/gorm"
)

var processSingleMediaFn = func(db *gorm.DB, media *models.Media) error {
	return scanner.ProcessSingleMedia(db, media)
}

func handleVideoRequest(
	w http.ResponseWriter,
	r *http.Request,
	db *gorm.DB,
	mediaName string,
	authenticateFn func(*models.Media, *gorm.DB, *http.Request) (bool, string, int, error),
	getCachePathFn func(albumID, mediaID int, filename string) string,
) {
	var mediaURLs []models.MediaURL
	if err := db.Model(&models.MediaURL{}).
		Preload("Media").
		Where("media_urls.media_name = ? AND media_urls.purpose = ?", mediaName, models.VideoWeb).
		Find(&mediaURLs).
		Error; err != nil || len(mediaURLs) == 0 || mediaURLs[0].Media == nil {

		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("not found"))
		return
	}

	if len(mediaURLs) > 1 {
		sanitizedMediaName := strings.Map(func(r rune) rune {
			if unicode.IsPrint(r) {
				return r
			}
			return -1
		}, mediaName)
		log.Warn("Multiple video web URLs found",
			"name", sanitizedMediaName,
			"count", len(mediaURLs),
			"using", mediaURLs[0],
		)
	}

	mediaURL := mediaURLs[0]
	var media = mediaURL.Media

	if success, response, status, err := authenticateFn(media, db, r); !success {
		if err != nil {
			log.Warn("got error authenticating video:",
				"error", err,
				"media ID", media.ID,
				"media path", media.Path)
		}
		w.WriteHeader(status)
		w.Write([]byte(response))
		return
	}

	var cachedPath string

	if mediaURL.Purpose == models.VideoWeb {
		// Use the provided cache path function
		cachedPath = getCachePathFn(int(media.AlbumID), int(mediaURL.MediaID), mediaURL.MediaName)
	} else {
		log.Error("Can not handle media_purpose for video",
			"purpose", mediaURL.Purpose,
			"expected", models.VideoWeb)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(internalServerError))
		return
	}

	// Context-aware processing function that respects client disconnections
	contextAwareProcessFn := func(db *gorm.DB, media *models.Media) error {
		// If the client disconnects during processing, this will be detected
		ctx := r.Context()

		// Create a channel to communicate processing completion
		done := make(chan error, 1)

		go func() {
			// Run the actual processing in a goroutine
			done <- processSingleMediaFn(db, media)
		}()

		// Wait for either processing to complete or context to be canceled
		select {
		case err := <-done:
			return err
		case <-ctx.Done():
			log.Warn("Client disconnected during video processing",
				"mediaID", media.ID,
				"reason", ctx.Err())
			return ctx.Err()
		}
	}

	if _, err := os.Stat(cachedPath); err != nil {
		if os.IsNotExist(err) {
			if err := contextAwareProcessFn(db, media); err != nil {
				log.Error("processing video not found in cache:",
					"error", err,
					"media ID", media.ID,
					"media path", media.Path)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(internalServerError))
				return
			}

			if _, err := os.Stat(cachedPath); err != nil {
				log.Error("video not found in cache after reprocessing:",
					"error", err,
					"media ID", media.ID,
					"media path", media.Path)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(internalServerError))
				return
			}
		} else {
			log.Error("cached video access error:",
				"error", err,
				"media ID", media.ID,
				"media path", media.Path)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(internalServerError))
			return
		}
	}

	w.Header().Set("Cache-Control", "private, max-age=86400, immutable")
	http.ServeFile(w, r, cachedPath)
}

func RegisterVideoRoutes(db *gorm.DB, router *mux.Router) {

	router.HandleFunc("/{name}", func(w http.ResponseWriter, r *http.Request) {
		mediaName := mux.Vars(r)["name"]

		handleVideoRequest(
			w, r, db, mediaName,
			authenticateMedia, // Real authentication
			func(albumID, mediaID int, filename string) string {
				return path.Join(utils.MediaCachePath(), strconv.Itoa(albumID), strconv.Itoa(mediaID), filename)
			},
		)
	})
}

func registerMockVideoRoutesForTesting(db *gorm.DB, router *mux.Router, tempCachePath string) {
	router.HandleFunc("/{name}", func(w http.ResponseWriter, r *http.Request) {
		mediaName := mux.Vars(r)["name"]

		// Use no-op auth and test cache path
		handleVideoRequest(
			w, r, db, mediaName,
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
