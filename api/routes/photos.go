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
	"golang.org/x/crypto/bcrypt"

	"github.com/viktorstrate/photoview/api/graphql/auth"
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

		user := auth.UserFromContext(r.Context())
		if user != nil {
			row := db.QueryRow("SELECT owner_id FROM album WHERE album.album_id = ?", media.AlbumId)
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
			// Check if photo is authorized with a share token
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

			// Validate share token password, if set
			if shareToken.Password != nil {
				tokenPassword := r.Header.Get("TokenPassword")

				if err := bcrypt.CompareHashAndPassword([]byte(*shareToken.Password), []byte(tokenPassword)); err != nil {
					if err == bcrypt.ErrMismatchedHashAndPassword {
						w.WriteHeader(http.StatusForbidden)
						w.Write([]byte("unauthorized"))
						return
					} else {
						w.WriteHeader(http.StatusInternalServerError)
						w.Write([]byte("internal server error"))
						return
					}
				}
			}

			if shareToken.AlbumID != nil && media.AlbumId != *shareToken.AlbumID {
				// Check child albums
				row := db.QueryRow(`
					WITH recursive child_albums AS (
						SELECT * FROM album WHERE parent_album = ?
						UNION ALL
						SELECT child.* FROM album child JOIN child_albums parent ON parent.album_id = child.parent_album
					)
					SELECT * FROM child_albums WHERE album_id = ?
				`, *shareToken.AlbumID, media.AlbumId)

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

			if shareToken.MediaID != nil && media_id != *shareToken.MediaID {
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte("unauthorized"))
				return
			}

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

		io.Copy(w, file)
	})
}
