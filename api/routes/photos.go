package routes

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"github.com/viktorstrate/photoview/api/graphql/auth"
	"github.com/viktorstrate/photoview/api/graphql/models"
)

func PhotoRoutes(db *sql.DB) chi.Router {
	router := chi.NewRouter()
	router.Get("/{name}", func(w http.ResponseWriter, r *http.Request) {
		image_name := chi.URLParam(r, "name")

		row := db.QueryRow("SELECT photo_url.purpose, photo.path, photo.photo_id, photo.album_id, photo_url.content_type FROM photo_url, photo WHERE photo_url.photo_name = ? AND photo_url.photo_id = photo.photo_id", image_name)

		var purpose models.PhotoPurpose
		var path string
		var content_type string
		var album_id int
		var photo_id int

		if err := row.Scan(&purpose, &path, &photo_id, &album_id, &content_type); err != nil {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("404"))
			return
		}

		user := auth.UserFromContext(r.Context())
		if user != nil {
			row := db.QueryRow("SELECT owner_id FROM album WHERE album.album_id = ?", album_id)
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

			if shareToken.AlbumID != nil && album_id != *shareToken.AlbumID {
				// Check child albums
				row := db.QueryRow(`
					WITH recursive child_albums AS (
						SELECT * FROM album WHERE parent_album = ?
						UNION ALL
						SELECT child.* FROM album child JOIN child_albums parent ON parent.album_id = child.parent_album
					)
					SELECT * FROM child_albums WHERE album_id = ?
				`, *shareToken.AlbumID, album_id)

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

		w.Header().Set("Content-Type", content_type)

		var file *os.File

		if purpose == models.PhotoThumbnail || purpose == models.PhotoHighRes {
			var err error
			file, err = os.Open(fmt.Sprintf("./image-cache/%d/%d/%s", album_id, photo_id, image_name))
			if err != nil {
				w.Write([]byte("Error: " + err.Error()))
				return
			}
		}

		if purpose == models.PhotoOriginal {
			var err error
			file, err = os.Open(path)
			if err != nil {
				w.Write([]byte("Error: " + err.Error()))
				return
			}
		}

		if stats, err := file.Stat(); err == nil {
			w.Header().Set("Content-Length", fmt.Sprintf("%d", stats.Size()))
		}

		io.Copy(w, file)
	})

	return router
}
