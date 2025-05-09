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

func RegisterVideoRoutes(db *gorm.DB, router *mux.Router) {

	router.HandleFunc("/{name}", func(w http.ResponseWriter, r *http.Request) {
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

		if success, response, status, err := authenticateMedia(media, db, r); !success {
			if err != nil {
				log.Warn("got error authenticating video:", err)
			}
			w.WriteHeader(status)
			w.Write([]byte(response))
			return
		}

		var cachedPath string

		if mediaURL.Purpose == models.VideoWeb {
			cachedPath = path.Join(utils.MediaCachePath(), strconv.Itoa(int(media.AlbumID)), strconv.Itoa(int(mediaURL.MediaID)), mediaURL.MediaName)
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
			}
		}

		http.ServeFile(w, r, cachedPath)
	})
}
