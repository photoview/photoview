package routes

import (
	"log"
	"net/http"
	"os"
	"path"
	"strconv"

	"github.com/gorilla/mux"
	"gorm.io/gorm"

	"github.com/viktorstrate/photoview/api/graphql/models"
	"github.com/viktorstrate/photoview/api/scanner"
)

func RegisterPhotoRoutes(db *gorm.DB, router *mux.Router) {

	router.HandleFunc("/{name}", func(w http.ResponseWriter, r *http.Request) {
		mediaName := mux.Vars(r)["name"]

		var mediaURL models.MediaURL
		result := db.Model(&models.MediaURL{}).Select("media_urls.*").Where("media_url.media_name = ?", mediaName).Joins("JOIN media ON media_urls.media_id = media.id").Scan(&mediaURL)
		if err := result.Error; err != nil {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("404"))
			return
		}

		media := &mediaURL.Media

		if success, response, status, err := authenticateMedia(media, db, r); !success {
			if err != nil {
				log.Printf("WARN: error authenticating photo: %s\n", err)
			}
			w.WriteHeader(status)
			w.Write([]byte(response))
			return
		}

		var cachedPath string

		if mediaURL.Purpose == models.PhotoThumbnail || mediaURL.Purpose == models.PhotoHighRes || mediaURL.Purpose == models.VideoThumbnail {
			cachedPath = path.Join(scanner.PhotoCache(), strconv.Itoa(int(media.AlbumID)), strconv.Itoa(int(mediaURL.MediaID)), mediaURL.MediaName)
		} else if mediaURL.Purpose == models.MediaOriginal {
			cachedPath = media.Path
		} else {
			log.Printf("ERROR: Can not handle media_purpose for photo: %s\n", mediaURL.Purpose)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("internal server error"))
			return
		}

		if _, err := os.Stat(cachedPath); os.IsNotExist((err)) {
			err := db.Transaction(func(tx *gorm.DB) error {
				if _, err = scanner.ProcessMedia(tx, media); err != nil {
					log.Printf("ERROR: processing image not found in cache: %s\n", err)
					return err
				}

				if _, err = os.Stat(cachedPath); err != nil {
					log.Printf("ERROR: after reprocessing image not found in cache: %s\n", err)
					return err
				}

				return nil
			})

			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("internal server error"))
				return
			}
		}

		// Allow caching the resource for 1 day
		w.Header().Set("Cache-Control", "private, max-age=86400, immutable")

		http.ServeFile(w, r, cachedPath)
	})
}
