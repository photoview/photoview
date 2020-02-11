package routes

import (
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/go-chi/chi"
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

		w.Header().Set("Content-Type", content_type)

		var file *os.File

		if purpose == models.PhotoThumbnail {
			var err error
			file, err = os.Open(fmt.Sprintf("./image-cache/%d/%d/%s", album_id, photo_id, image_name))
			if err != nil {
				w.Write([]byte("Error: " + err.Error()))
				return
			}
		}

		if purpose == models.PhotoHighRes {
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
