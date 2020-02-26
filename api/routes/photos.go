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

	"github.com/viktorstrate/photoview/api/graphql/auth"
	"github.com/viktorstrate/photoview/api/graphql/models"
	"github.com/viktorstrate/photoview/api/scanner"
)

func RegisterPhotoRoutes(db *sql.DB, router *mux.Router) {

	router.HandleFunc("/{name}", func(w http.ResponseWriter, r *http.Request) {
		image_name := mux.Vars(r)["name"]

		row := db.QueryRow("SELECT photo_url.purpose, photo_url.content_type, photo_url.photo_id FROM photo_url, photo WHERE photo_url.photo_name = ? AND photo_url.photo_id = photo.photo_id", image_name)

		var purpose models.PhotoPurpose
		var content_type string
		var photo_id int

		if err := row.Scan(&purpose, &content_type, &photo_id); err != nil {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("404"))
			return
		}

		row = db.QueryRow("SELECT * FROM photo WHERE photo_id = ?", photo_id)
		photo, err := models.NewPhotoFromRow(row)
		if err != nil {
			log.Printf("WARN: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("internal server error"))
		}

		user := auth.UserFromContext(r.Context())
		if user != nil {
			row := db.QueryRow("SELECT owner_id FROM album WHERE album.album_id = ?", photo.AlbumId)
			var owner_id int

			if err := row.Scan(&owner_id); err != nil {
				log.Printf("WARN: %s", err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("internal server error"))
				return
			}

			if owner_id != user.UserID {
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte("invalid credentials"))
				return
			}
		} else {

			token := r.URL.Query().Get("token")
			if token == "" {
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte("unauthorized"))
				return
			}

			row := db.QueryRow("SELECT * FROM share_token WHERE value = ?", token)

			shareToken, err := models.NewShareTokenFromRow(row)
			if err != nil {
				log.Printf("WARN: %s", err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("internal server error"))
				return
			}

			if shareToken.AlbumID != nil && photo.AlbumId != *shareToken.AlbumID {
				// Check child albums
				row := db.QueryRow(`
					WITH recursive child_albums AS (
						SELECT * FROM album WHERE parent_album = ?
						UNION ALL
						SELECT child.* FROM album child JOIN child_albums parent ON parent.album_id = child.parent_album
					)
					SELECT * FROM child_albums WHERE album_id = ?
				`, *shareToken.AlbumID, photo.AlbumId)

				_, err := models.NewAlbumFromRow(row)
				if err != nil {
					if err == sql.ErrNoRows {
						w.WriteHeader(http.StatusForbidden)
						w.Write([]byte("unauthorized"))
						return
					}
					log.Printf("WARN: %s", err)
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte("internal server error"))
					return
				}
			}

			if shareToken.PhotoID != nil && photo_id != *shareToken.PhotoID {
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte("unauthorized"))
				return
			}

		}

		var cachedPath string
		var file *os.File = nil

		if purpose == models.PhotoThumbnail || purpose == models.PhotoHighRes {
			cachedPath = path.Join(scanner.PhotoCache(), strconv.Itoa(photo.AlbumId), strconv.Itoa(photo_id), image_name)
		}

		if purpose == models.PhotoOriginal {
			cachedPath = photo.Path
		}

		file, err = os.Open(cachedPath)
		if err != nil {
			if os.IsNotExist(err) {
				tx, err := db.Begin()
				if err != nil {
					log.Printf("ERROR: %s\n", err)
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte("internal server error"))
					return
				}

				err = scanner.ProcessPhoto(tx, photo)
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

		io.Copy(w, file)
	})
}
