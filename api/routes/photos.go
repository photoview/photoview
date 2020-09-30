package routes

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/viktorstrate/photoview/api/graphql/models"
	"github.com/viktorstrate/photoview/api/scanner"
)

func RegisterPhotoRoutes(db *sql.DB, router *mux.Router) {

	router.HandleFunc("/{name}", func(w http.ResponseWriter, r *http.Request) {
		media_name := mux.Vars(r)["name"]

		row := db.QueryRow("SELECT media_url.purpose, media_url.content_type, media_url.media_id FROM media_url, media WHERE media_url.media_name = ? AND media_url.media_id = media.media_id", media_name)

		var purpose models.MediaPurpose
		var content_type string
		var media_id int

		if err := row.Scan(&purpose, &content_type, &media_id); err != nil {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("404"))
			return
		}

		row = db.QueryRow("SELECT * FROM media WHERE media_id = ?", media_id)
		media, err := models.NewMediaFromRow(row)
		if err != nil {
			log.Printf("WARN: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("internal server error"))
		}

		if success, response, status, err := authenticateMedia(media, db, r); !success {
			if err != nil {
				log.Printf("WARN: error authenticating photo: %s\n", err)
			}
			w.WriteHeader(status)
			w.Write([]byte(response))
			return
		}

		var cachedPath string
		var file *os.File = nil

		if purpose == models.PhotoThumbnail || purpose == models.PhotoHighRes || purpose == models.VideoThumbnail {
			cachedPath = path.Join(scanner.PhotoCache(), strconv.Itoa(media.AlbumId), strconv.Itoa(media_id), media_name)
		} else if purpose == models.MediaOriginal {
			cachedPath = media.Path
		} else {
			log.Printf("ERROR: Can not handle media_purpose for photo: %s\n", purpose)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("internal server error"))
			return
		}

		file, err = os.Open(cachedPath)
		defer file.Close()
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
					log.Printf("ERROR: processing image not found in cache: %s\n", err)
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte("internal server error"))
					tx.Rollback()
					return
				}

				file, err = os.Open(cachedPath)
				if err != nil {
					log.Printf("ERROR: after reprocessing image not found in cache: %s\n", err)
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte("internal server error"))
					tx.Rollback()
					return
				}

				tx.Commit()
			}
		}

		w.Header().Set("Content-Type", content_type)
		if stats, err := file.Stat(); err == nil {
			w.Header().Set("Content-Length", fmt.Sprintf("%d", stats.Size()))
		}

		// Allow caching the resource for 1 day
		w.Header().Set("Cache-Control", "private, max-age=86400, immutable")

		io.Copy(w, file)
	})
}
