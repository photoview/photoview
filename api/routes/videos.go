package routes

import (
	"log"
	"net/http"
	"os"
	"path"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/viktorstrate/photoview/api/graphql/models"
	"github.com/viktorstrate/photoview/api/scanner"
	"gorm.io/gorm"
)

func RegisterVideoRoutes(db *gorm.DB, router *mux.Router) {

	router.HandleFunc("/{name}", func(w http.ResponseWriter, r *http.Request) {
		media_name := mux.Vars(r)["name"]

		row := db.QueryRow("SELECT media_url.* FROM media_url JOIN media ON media_url.media_id = media.media_id WHERE media_url.media_name = ?", media_name)
		mediaURL, err := models.NewMediaURLFromRow(row)

		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("404"))
			return
		}

		row = db.QueryRow("SELECT * FROM media WHERE media_id = ?", mediaURL.MediaId)
		media, err := models.NewMediaFromRow(row)
		if err != nil {
			log.Printf("WARN: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("internal server error"))
		}

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
			cachedPath = path.Join(scanner.PhotoCache(), strconv.Itoa(media.AlbumId), strconv.Itoa(mediaURL.MediaId), mediaURL.MediaName)
		} else {
			log.Printf("ERROR: Can not handle media_purpose for video: %s\n", mediaURL.Purpose)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("internal server error"))
			return
		}

		_, err = os.Stat(cachedPath)
		if err != nil {
			if os.IsNotExist(err) {
				tx, err := db.Begin()
				if err != nil {
					log.Printf("ERROR: %s\n", err)
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte("internal server error"))
					return
				}

				_, err = scanner.ProcessMedia(tx, media)
				if err != nil {
					log.Printf("ERROR: processing video not found in cache: %s\n", err)
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte("internal server error"))
					tx.Rollback()
					return
				}

				_, err = os.Stat(cachedPath)
				if err != nil {
					log.Printf("ERROR: after reprocessing video not found in cache: %s\n", err)
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte("internal server error"))
					tx.Rollback()
					return
				}

				tx.Commit()
			}
		}

		http.ServeFile(w, r, cachedPath)
	})
}
