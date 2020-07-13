package routes

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/viktorstrate/photoview/api/graphql/models"
	"github.com/viktorstrate/photoview/api/scanner"
)

func RegisterVideoRoutes(db *sql.DB, router *mux.Router) {

	router.HandleFunc("/{name}", func(w http.ResponseWriter, r *http.Request) {
		media_name := mux.Vars(r)["name"]

		row := db.QueryRow("SELECT media_url.purpose, media_url.media_id FROM media_url, media WHERE media_url.media_name = ? AND media_url.media_id = media.media_id", media_name)

		var purpose models.MediaPurpose
		var media_id int

		if err := row.Scan(&purpose, &media_id); err != nil {
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
				log.Printf("WARN: error authenticating video: %s\n", err)
			}
			w.WriteHeader(status)
			w.Write([]byte(response))
			return
		}

		log.Printf("Video cookies: %d", len(r.Cookies()))

		var cachedPath string

		if purpose == models.VideoWeb {
			cachedPath = path.Join(scanner.PhotoCache(), strconv.Itoa(media.AlbumId), strconv.Itoa(media_id), media_name)
		} else {
			log.Printf("ERROR: Can not handle media_purpose for video: %s\n", purpose)
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
