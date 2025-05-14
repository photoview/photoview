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
		Select("media_urls.*, Media.*").
		Preload("Media").
		Joins("Media").
		Where("media_urls.media_name = ? AND media_urls.purpose = ?", mediaName, models.VideoWeb).
		Find(&mediaURLs).
		Error; err != nil || len(mediaURLs) == 0 || mediaURLs[0].Media == nil {

		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404"))
		return
	}

	if len(mediaURLs) > 1 {
		sanitizedMediaName := strings.Map(func(r rune) rune {
			if unicode.IsPrint(r) {
				return r
			}
			return -1
		}, mediaName)
		log.Warn("Multiple video web URLs found for name", sanitizedMediaName, "count", len(mediaURLs), "using", mediaURLs[0])
	}

	mediaURL := mediaURLs[0]
	var media = mediaURL.Media

	// Use the provided authentication function
	if success, response, status, err := authenticateFn(media, db, r); !success {
		if err != nil {
			log.Warn("got error authenticating video:", err)
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
		log.Error("Can not handle media_purpose for video:", mediaURL.Purpose)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(internalServerError))
		return
	}

	if _, err := os.Stat(cachedPath); err != nil {
		if os.IsNotExist(err) {
			if err := scanner.ProcessSingleMedia(db, media); err != nil {
				log.Error("processing video not found in cache:", err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(internalServerError))
				return
			}

			if _, err := os.Stat(cachedPath); err != nil {
				log.Error("video not found in cache after reprocessing:", err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(internalServerError))
				return
			}
		} else {
			log.Error("cached video access error:", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(internalServerError))
			return
		}
	}

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
				return true, "success", http.StatusAccepted, nil
			},
			// Use test cache path
			func(albumID, mediaID int, filename string) string {
				return path.Join(tempCachePath, strconv.Itoa(albumID), strconv.Itoa(mediaID), filename)
			},
		)
	})
}
