package routes

import (
	"context"
	"net/http"
	"os"
	"path"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/log"
	"github.com/photoview/photoview/api/scanner"
	"github.com/photoview/photoview/api/utils"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

var processSingleMediaFn = func(ctx context.Context, db *gorm.DB, media *models.Media) error {
	return scanner.ProcessSingleMedia(ctx, db, media)
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
		Order("created_at DESC").
		Find(&mediaURLs).
		Error; err != nil || len(mediaURLs) == 0 || mediaURLs[0].Media == nil {

		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("not found"))
		return
	}

	if len(mediaURLs) > 1 {
		log.Warn(r.Context(), "Multiple video web URLs found",
			"name", mediaName,
			"count", len(mediaURLs),
			"using", mediaURLs[0],
		)
	}

	mediaURL := mediaURLs[0]
	var media = mediaURL.Media

	if success, response, status, err := authenticateFn(media, db, r); !success {
		if err != nil {
			log.Warn(r.Context(), "got error authenticating video",
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
		log.Error(r.Context(), "Can not handle media_purpose for video",
			"purpose", mediaURL.Purpose,
			"expected", models.VideoWeb)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(internalServerError))
		return
	}

	if _, err := os.Stat(cachedPath); err != nil {
		if !os.IsNotExist(err) {
			log.Error(r.Context(), "cached video access error",
				"error", err,
				"media ID", media.ID,
				"media path", media.Path)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(internalServerError))
			return
		}

		if err := processSingleMediaFn(r.Context(), db, media); err != nil {
			// Check if error was due to context cancellation
			if r.Context().Err() != nil && errors.Is(r.Context().Err(), context.Canceled) {
				log.Warn(r.Context(), "video processing cancelled due to client disconnect",
					"mediaID", media.ID,
					"reason", r.Context().Err())
				return // Don't send response if client disconnected
			}

			log.Error(r.Context(), "processing video not found in cache",
				"error", err,
				"media ID", media.ID,
				"media path", media.Path)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(internalServerError))
			return
		}

		if _, err := os.Stat(cachedPath); err != nil {
			log.Error(r.Context(), "video not found in cache after reprocessing",
				"error", err,
				"media ID", media.ID,
				"media path", media.Path)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(internalServerError))
			return
		}
	}

	w.Header().Set("Cache-Control", "private, max-age=86400, immutable")
	w.Header().Set("Content-Type", mediaURL.ContentType)
	http.ServeFile(w, r, cachedPath)
}

func generateCacheFilename(albumID, mediaID int, filename string) string {
	return path.Join(utils.MediaCachePath(), strconv.Itoa(albumID), strconv.Itoa(mediaID), filename)
}

func RegisterVideoRoutes(db *gorm.DB, router *mux.Router) {

	router.HandleFunc("/{name}", func(w http.ResponseWriter, r *http.Request) {
		mediaName := mux.Vars(r)["name"]
		handleVideoRequest(w, r, db, mediaName, authenticateMedia, generateCacheFilename)
	})
}
