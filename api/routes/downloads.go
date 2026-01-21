package routes

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/photoview/photoview/api/database/drivers"
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/spf13/afero"
	"gorm.io/gorm"
)

func RegisterDownloadRoutes(db *gorm.DB, fileFs afero.Fs, cacheFs afero.Fs, router *mux.Router) {
	router.HandleFunc("/album/{album_id}/{media_purpose}", func(w http.ResponseWriter, r *http.Request) {
		albumID := mux.Vars(r)["album_id"]
		mediaPurpose := mux.Vars(r)["media_purpose"]
		mediaPurposeList := strings.SplitN(mediaPurpose, ",", 10)

		var album models.Album
		if err := db.Find(&album, albumID).Error; err != nil {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("404"))
			return
		}

		if success, response, status, err := authenticateAlbum(&album, db, r); !success {
			if err != nil {
				log.Printf("WARN: error authenticating album for download: %v\n", err)
			}
			w.WriteHeader(status)
			w.Write([]byte(response))
			return
		}

		var mediaWhereQuery string
		if drivers.POSTGRES.MatchDatabase(db) {
			mediaWhereQuery = "\"Media\".album_id = ?"
		} else {
			mediaWhereQuery = "Media.album_id = ?"
		}

		var mediaURLs []*models.MediaURL
		if err := db.Joins("Media").Where(mediaWhereQuery, album.ID).Where("media_urls.purpose IN (?)", mediaPurposeList).Find(&mediaURLs).Error; err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(internalServerError))
			return
		}

		if len(mediaURLs) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("no media found"))
			return
		}

		// Do not allow caching
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Content-Type", "application/zip")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.zip\"", album.Title))

		zipWriter := zip.NewWriter(w)

		for _, media := range mediaURLs {
			zipFile, err := zipWriter.Create(fmt.Sprintf("%s/%s", album.Title, media.MediaName))
			if err != nil {
				log.Printf("ERROR: Failed to create a file in zip, when downloading album (%d): %v\n", album.ID, err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(internalServerError))
				return
			}

			filePath, err := media.CachedPath()
			if err != nil {
				log.Printf("ERROR: Failed to get mediaURL cache path, when downloading album (%d): %v\n", album.ID, err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(internalServerError))
				return
			}

			fs := fileFs
			if media.Purpose != models.MediaOriginal {
				fs = cacheFs
			}

			fileData, err := fs.Open(filePath)
			if err != nil {
				log.Printf("ERROR: Failed to open file to include in zip, when downloading album (%d): %v\n", album.ID, err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(internalServerError))
				return
			}

			_, err = io.Copy(zipFile, fileData)
			if err != nil {
				_ = fileData.Close()
				log.Printf("ERROR: Failed to copy file data, when downloading album (%d): %v\n", album.ID, err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(internalServerError))
				return
			}

			// Close the file directly after copying (instead of deferring) to avoid too many open files
			if err := fileData.Close(); err != nil {
				log.Printf("WARN: Failed to close file after zip write: %v\n", err)
			}
		}

		// close the zip Writer to flush the contents to the ResponseWriter
		zipWriter.Close()
	})
}
