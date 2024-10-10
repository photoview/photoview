package routes

import (
	"log"
	"net/http"
	"path"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/utils"
	"gorm.io/gorm"
)

func RegisterVideoRoutes(db *gorm.DB, router *mux.Router) {

	router.HandleFunc("/{name}", func(w http.ResponseWriter, r *http.Request) {
		mediaName := mux.Vars(r)["name"]

		var mediaURL models.MediaURL
		result := db.
			Model(&models.MediaURL{}).
			Select("media_urls.*").
			Joins("Media").
			Where("media_urls.media_name = ?", mediaName).
			Find(&mediaURL)
		if err := result.Error; err != nil {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("404"))
			return
		}

		var media = mediaURL.Media

		if success, response, status, err := authenticateMedia(media, db, r); !success {
			if err != nil {
				log.Printf("WARN: error authenticating video: %s\n", err)
			}
			w.WriteHeader(status)
			w.Write([]byte(response))
			return
		}

		var cachedPath string

		if mediaURL.Purpose == models.VideoWeb {
			cachedPath = path.Join(utils.MediaCachePath(), strconv.Itoa(int(media.AlbumID)),
				strconv.Itoa(int(mediaURL.MediaID)), mediaURL.MediaName)
		} else {
			log.Printf("ERROR: Can not handle media_purpose for video: %s\n", mediaURL.Purpose)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(internalServerError))
			return
		}

		if isError := addMediaToCache(cachedPath, db, media, "video", w); isError {
			return
		}

		http.ServeFile(w, r, cachedPath)
	})
}
