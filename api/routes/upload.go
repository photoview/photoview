package routes

import (
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"gorm.io/gorm"

	"github.com/photoview/photoview/api/graphql/auth"
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/log"
	"github.com/photoview/photoview/api/scanner/media_type"
	"github.com/photoview/photoview/api/scanner/scanner_queue"
	"github.com/photoview/photoview/api/utils"
)

// maxUploadBytes bounds the total size of a single upload request.
const maxUploadBytes = 2 << 30 // 2 GiB

// maxUploadMemoryBytes is the in-memory threshold for multipart parsing;
// parts larger than this spill to temp files automatically.
const maxUploadMemoryBytes = 32 << 20 // 32 MiB

type uploadFileResult struct {
	Path   string `json:"path"`
	Status string `json:"status"` // "ok" | "rejected" | "error"
	Reason string `json:"reason,omitempty"`
}

type uploadResponse struct {
	Results []uploadFileResult `json:"results"`
}

// addAlbumToQueue is a seam over scanner_queue.AddAlbumToQueue so tests can
// avoid touching the real, process-global scanner queue (which starts a
// background worker goroutine on first use).
var addAlbumToQueue = scanner_queue.AddAlbumToQueue

// RegisterUploadRoutes registers the endpoint used to upload media into an
// album a user owns. Unlike the read-only photo/video/download routes, this
// is never reachable via a share token — the caller must be a logged in
// user holding at least Upload-level access on the target album (or an
// admin).
func RegisterUploadRoutes(db *gorm.DB, router *mux.Router) {
	router.HandleFunc("/{albumId}", func(w http.ResponseWriter, r *http.Request) {
		albumID, err := strconv.Atoi(mux.Vars(r)["albumId"])
		if err != nil {
			http.Error(w, "invalid album id", http.StatusBadRequest)
			return
		}

		user := auth.UserFromContext(r.Context())
		if user == nil {
			http.Error(w, "unauthorized", http.StatusForbidden)
			return
		}

		// A missing album and a forbidden one return the same status and
		// body, so probing album ids can't be used to enumerate which ones
		// exist.
		var album models.Album
		if err := db.First(&album, albumID).Error; err != nil {
			http.Error(w, "unauthorized", http.StatusForbidden)
			return
		}

		canUpload, err := user.HasAlbumLevel(db, &album, models.AlbumPermissionLevelUpload)
		if err != nil {
			log.Error(r.Context(), "error checking upload permission", "error", err)
			http.Error(w, internalServerError, http.StatusInternalServerError)
			return
		}
		if !canUpload {
			http.Error(w, "unauthorized", http.StatusForbidden)
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, maxUploadBytes)
		if err := r.ParseMultipartForm(maxUploadMemoryBytes); err != nil {
			http.Error(w, "upload too large or malformed", http.StatusBadRequest)
			return
		}
		defer r.MultipartForm.RemoveAll()

		results := make([]uploadFileResult, 0, len(r.MultipartForm.File))
		anyStored := false

		for relPath, headers := range r.MultipartForm.File {
			for _, header := range headers {
				result := saveOneUploadedFile(&album, relPath, header)
				results = append(results, result)
				if result.Status == "ok" {
					anyStored = true
				}
			}
		}

		if anyStored {
			// One rescan of the whole target album picks up every stored
			// file/new sub-folder above; no need to queue per-file.
			if err := addAlbumToQueue(&album); err != nil {
				log.Error(r.Context(), "failed to queue album for scanning after upload", "error", err)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(uploadResponse{Results: results}); err != nil {
			log.Error(r.Context(), "failed to encode upload response", "error", err)
		}
		// Registering only "POST" here would make gorilla/mux respond to an
		// OPTIONS preflight with 405 *before* the CORS middleware (which
		// short-circuits OPTIONS requests itself) ever runs, since mux only
		// invokes Router.Use() middleware for requests that match a route.
	}).Methods("POST", "OPTIONS")
}

var errUploadPathSymlink = errors.New("path crosses a symbolic link")

// prepareUploadDir creates the directories cleanRel needs below albumPath,
// rejecting any component that already exists as a symlink. Containment
// checks on the joined path are purely lexical, so a symlink sitting inside
// the album (say "trip" -> /etc) passes them while the write itself lands
// wherever the link points. Uploading into a symlinked sub-album is still
// possible by selecting that album directly, where its own resolved path is
// the root this walks from.
func prepareUploadDir(albumPath string, cleanRel string) error {
	relDir := filepath.Dir(cleanRel)
	if relDir == "." {
		return nil
	}

	dir := albumPath
	for _, part := range strings.Split(relDir, string(filepath.Separator)) {
		dir = filepath.Join(dir, part)

		info, err := os.Lstat(dir)
		if err != nil {
			if !os.IsNotExist(err) {
				return err
			}
			if err := os.Mkdir(dir, 0o755); err != nil {
				return err
			}
			continue
		}

		if info.Mode()&os.ModeSymlink != 0 {
			return errUploadPathSymlink
		}
		if !info.IsDir() {
			return errors.New("path component is not a directory")
		}
	}

	return nil
}

func saveOneUploadedFile(album *models.Album, relPath string, header *multipart.FileHeader) uploadFileResult {
	cleanRel, err := utils.SanitizeRelativePath(relPath)
	if err != nil {
		return uploadFileResult{Path: relPath, Status: "rejected", Reason: err.Error()}
	}

	destPath := filepath.Join(album.Path, cleanRel)
	if !utils.IsSubPath(album.Path, destPath) {
		return uploadFileResult{Path: relPath, Status: "rejected", Reason: "path escapes album directory"}
	}

	if err := prepareUploadDir(album.Path, cleanRel); err != nil {
		if errors.Is(err, errUploadPathSymlink) {
			return uploadFileResult{Path: relPath, Status: "rejected", Reason: err.Error()}
		}
		return uploadFileResult{Path: relPath, Status: "error", Reason: "could not create directory"}
	}

	src, err := header.Open()
	if err != nil {
		return uploadFileResult{Path: relPath, Status: "error", Reason: "could not read upload"}
	}
	defer src.Close()

	// Write to a temp name first, so a rejected or half-written file is never
	// visible under its real name where a background scan could see it.
	tmpPath := destPath + ".photoview-upload-tmp"
	dst, err := os.OpenFile(tmpPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return uploadFileResult{Path: relPath, Status: "error", Reason: "could not create file"}
	}

	if _, err := io.Copy(dst, src); err != nil {
		dst.Close()
		os.Remove(tmpPath)
		return uploadFileResult{Path: relPath, Status: "error", Reason: "write failed"}
	}
	// A delayed write failure (e.g. a full disk) can surface only here, not
	// from io.Copy - check it rather than accepting a possibly-truncated file.
	if err := dst.Close(); err != nil {
		os.Remove(tmpPath)
		return uploadFileResult{Path: relPath, Status: "error", Reason: "could not finalize write"}
	}

	// Same acceptance rule the scanner itself uses (MIME sniffed from
	// content, not the client-declared type or file extension).
	if !media_type.GetMediaType(tmpPath).IsSupported() {
		os.Remove(tmpPath)
		return uploadFileResult{Path: relPath, Status: "rejected", Reason: "unsupported file type"}
	}

	// os.Rename would silently overwrite an existing file at destPath -
	// reject the upload instead of clobbering a library file already there.
	// os.Link refuses an existing destination itself, so unlike a
	// stat-then-rename guard it leaves no window for a second upload of the
	// same name to land in between and be overwritten.
	linkErr := os.Link(tmpPath, destPath)
	if linkErr == nil {
		os.Remove(tmpPath)
		return uploadFileResult{Path: relPath, Status: "ok"}
	}
	if os.IsExist(linkErr) {
		os.Remove(tmpPath)
		return uploadFileResult{Path: relPath, Status: "rejected", Reason: "a file with that name already exists"}
	}

	// Not every filesystem a library can live on supports hard links (exFAT
	// and some network mounts don't), so fall back to the checked rename
	// there rather than refusing the upload outright.
	if _, err := os.Stat(destPath); err == nil {
		os.Remove(tmpPath)
		return uploadFileResult{Path: relPath, Status: "rejected", Reason: "a file with that name already exists"}
	} else if !os.IsNotExist(err) {
		os.Remove(tmpPath)
		return uploadFileResult{Path: relPath, Status: "error", Reason: "could not check destination path"}
	}

	if err := os.Rename(tmpPath, destPath); err != nil {
		os.Remove(tmpPath)
		return uploadFileResult{Path: relPath, Status: "error", Reason: "could not finalize file"}
	}

	return uploadFileResult{Path: relPath, Status: "ok"}
}
