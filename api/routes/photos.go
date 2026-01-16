package routes

import (
	"io"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/spf13/afero"
	"gorm.io/gorm"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/log"
	"github.com/photoview/photoview/api/scanner"
)

func RegisterPhotoRoutes(db *gorm.DB, fs afero.Fs, router *mux.Router) {

	router.HandleFunc("/{name}", func(w http.ResponseWriter, r *http.Request) {
		mediaName := mux.Vars(r)["name"]

		var mediaURL models.MediaURL
		result := db.Model(&models.MediaURL{}).Joins("Media").Select("media_urls.*").Where("media_urls.media_name = ?", mediaName).Scan(&mediaURL)
		if err := result.Error; err != nil {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("404"))
			return
		}

		media := mediaURL.Media
		if media == nil {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("404 - Media not found"))
			return
		}

		if success, response, status, err := authenticateMedia(media, db, r); !success {
			if err != nil {
				log.Warn(r.Context(), "Unauthorized access to photo", "reason", err)
			}
			w.WriteHeader(status)
			w.Write([]byte(response))
			return
		}

		cachedPath, err := mediaURL.CachedPath()
		if err != nil {
			log.Error(r.Context(), "error getting cached path for media URL", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(internalServerError))
			return
		}

		if _, err := fs.Stat(cachedPath); os.IsNotExist(err) {
			// err := db.Transaction(func(tx *gorm.DB) error {
			if err = scanner.ProcessSingleMediaFunc(r.Context(), db, fs, media); err != nil {
				log.Error(r.Context(), "processing image not found in cache",
					"media_cache_path", cachedPath,
					"error", err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(internalServerError))
				return
			}

			if _, err = fs.Stat(cachedPath); err != nil {
				log.Error(r.Context(), "after reprocessing image not found in cache",
					"media_cache_path", cachedPath,
					"error", err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(internalServerError))
				return
			}
		}

		// Allow caching the resource
		w.Header().Set("Cache-Control", "private, max-age=31536000, immutable")
		if mediaURL.ContentType != "" {
			w.Header().Set("Content-Type", mediaURL.ContentType)
		}

		file, err := fs.Open(cachedPath)
		if err != nil {
			log.Error(r.Context(), "error opening cached media",
				"media_cache_path", cachedPath,
				"error", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(internalServerError))
			return
		}
		defer file.Close()

		stat, err := file.Stat()
		if err != nil {
			log.Error(r.Context(), "error statting cached media",
				"media_cache_path", cachedPath,
				"error", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(internalServerError))
			return
		}

		if seeker, ok := file.(io.ReadSeeker); ok {
			http.ServeContent(w, r, stat.Name(), stat.ModTime(), seeker)
			return
		}

		http.ServeContent(w, r, stat.Name(), stat.ModTime(), file)
	})
}
