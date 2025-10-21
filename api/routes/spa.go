package routes

import (
	"fmt"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/photoview/photoview/api/log"
)

// SpaHandler implements the http.Handler interface, so we can use it
// to respond to HTTP requests. The path to the static directory and
// path to the index file within that static directory are used to
// serve the SPA in the given static directory.
type SpaHandler struct {
	staticPath string
	indexPath  string
}

func NewSpaHandler(staticPath string, indexPath string) (SpaHandler, error) {
	// Keeping this validation here just because we need the absolute static path anyway
	staticPathAbs, err := filepath.Abs(staticPath)
	if err != nil {
		return SpaHandler{}, fmt.Errorf("static path %s is not valid: %w", staticPath, err)
	}

	if err := validPath(staticPathAbs, true); err != nil {
		return SpaHandler{}, fmt.Errorf("static path %s validation error: %w", staticPath, err)
	}
	if err := validPath(filepath.Join(staticPath, indexPath), false); err != nil {
		return SpaHandler{}, fmt.Errorf("index path %s validation error: %w", indexPath, err)
	}

	return SpaHandler{
		indexPath:  indexPath,
		staticPath: staticPathAbs,
	}, nil
}

// ServeHTTP inspects the URL path to locate a file within the static dir
// on the SPA handler. If a file is found, it will be served. If not, the
// file located at the index path on the SPA handler will be served. This
// is suitable behavior for serving an SPA (single page application).
// Pre-compressed files (.br, .zst, .gz) are served if the client supports
// them, otherwise the original file is served.
func (h SpaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	relPath := strings.TrimPrefix(filepath.Clean(r.URL.Path), "/")
	fullPath := filepath.Join(h.staticPath, relPath)

	r = r.WithContext(log.WithAttrs(r.Context(), "static_path", h.staticPath, "requested_path", r.URL.Path))

	// Special case: root path should serve index.html
	if relPath == "" {
		h.serveIndexHTML(w, r)
		return
	}

	absPath, err := filepath.Abs(fullPath)
	if err != nil {
		log.Error(r.Context(), "error building absolute path", "error", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	rel, err := filepath.Rel(h.staticPath, absPath)
	if err != nil || strings.Contains(rel, "..") {
		log.Error(r.Context(), "requested path is outside of static path", "error", err)
		http.Error(w, "Invalid request URI", http.StatusBadRequest)
		return
	}

	// Check if the original file exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		// File doesn't exist - let serveOriginal handle SPA routing
		h.serveOriginal(w, r, fullPath, relPath)
		return
	}

	// Don't compress already-compressed files
	if isCompressedFormat(filepath.Ext(absPath)) {
		h.serveOriginal(w, r, fullPath, relPath)
		return
	}
	// Try to serve pre-compressed file
	if h.servePrecompressedFile(w, r, fullPath, relPath) {
		return
	}

	// Fallback to original SPA handler logic
	h.serveOriginal(w, r, fullPath, relPath)
}

func (h SpaHandler) serveOriginal(w http.ResponseWriter, r *http.Request, fullPath string, relPath string) {
	// Check whether a file exists at the given path
	_, err := os.Stat(fullPath)
	if os.IsNotExist(err) {
		// File does not exist, serve index.html (SPA routing)
		h.serveIndexHTML(w, r)
		return
	}
	if err != nil {
		// If we got an error (that wasn't that the file doesn't exist) stating the file,
		// return a 500 internal server error and stop
		log.Error(r.Context(), "Error stating file, requested by client", "file_path", fullPath, "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set cache headers
	h.setCacheHeaders(w, relPath)

	// Use http.FileServer to serve the static file with no compression
	http.FileServer(http.Dir(h.staticPath)).ServeHTTP(w, r)
}

// servePrecompressedFile attempts to serve a pre-compressed variant of the file
// Returns true if a pre-compressed file was served, false otherwise
func (h SpaHandler) servePrecompressedFile(w http.ResponseWriter, r *http.Request, fullPath string, relPath string) bool {
	// Parse Accept-Encoding header
	acceptEncoding := filterZeroQuality(strings.ToLower(r.Header.Get("Accept-Encoding")))
	if acceptEncoding == "" {
		return false
	}

	// Check for pre-compressed variants in order of preference (br > zstd > gzip)
	encodings := []struct {
		name      string
		extension string
	}{
		{"br", ".br"},
		{"zstd", ".zst"},
		{"gzip", ".gz"},
	}

	for _, enc := range encodings {
		if !strings.Contains(acceptEncoding, enc.name) {
			continue
		}

		precompressedPath := fullPath + enc.extension
		if stat, err := os.Stat(precompressedPath); err != nil || stat.IsDir() {
			continue
		}

		// Detect Content-Type from the ORIGINAL file extension, not the compressed one
		contentType := mime.TypeByExtension(filepath.Ext(fullPath))
		if contentType != "" {
			w.Header().Set("Content-Type", contentType)
		}

		w.Header().Set("Content-Encoding", enc.name)
		w.Header().Add("Vary", "Accept-Encoding")
		// Set cache headers based on request path
		h.setCacheHeaders(w, relPath)

		// Serve pre-compressed file
		http.ServeFile(w, r, precompressedPath)
		return true
	}

	return false
}

// serveIndexHTML serves index.html with pre-compressed priority
func (h SpaHandler) serveIndexHTML(w http.ResponseWriter, r *http.Request) {
	indexPath := filepath.Join(h.staticPath, h.indexPath)

	// Try to serve pre-compressed index.html first
	// Use full and relative paths to index.html instead of originally requested paths
	if h.servePrecompressedFile(w, r, indexPath, h.indexPath) {
		return
	}

	// Fallback to uncompressed index.html
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		// Index file doesn't exist - this is a serious configuration error, not a regular 404
		log.Error(r.Context(), "Error: index.html not found", "index_path", indexPath)
		http.Error(w, "Application index file not found", http.StatusInternalServerError)
		return
	}

	h.setCacheHeaders(w, h.indexPath)
	http.ServeFile(w, r, indexPath)
}

// setCacheHeaders sets appropriate cache headers based on the request path
func (h SpaHandler) setCacheHeaders(w http.ResponseWriter, relPath string) {
	if strings.HasPrefix(relPath, "assets/") {
		// Long-term cache for fingerprinted assets
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	} else {
		// Short cache with revalidation for other files
		w.Header().Set("Cache-Control", "public, max-age=3600, must-revalidate")
	}
}

// validPath validates the given path string.
//
// Parameters:
//   - pathStr: The path to validate.
//   - isDir: A boolean indicating whether the path is expected to be a directory (true) or a file (false).
//
// Returns:
//   - error: An error if the path is invalid, or nil if the path is valid.
func validPath(pathStr string, isDir bool) error {
	pathAbs, err := filepath.Abs(pathStr)
	if err != nil {
		return fmt.Errorf("path %s is not valid: %w", pathStr, err)
	}

	stat, err := os.Stat(pathAbs)
	if os.IsNotExist(err) {
		return fmt.Errorf("path %s does not exist", pathAbs)
	}
	if os.IsPermission(err) {
		return fmt.Errorf("no permission to access path %s", pathAbs)
	}
	if err != nil {
		return fmt.Errorf("error accessing path %s: %w", pathAbs, err)
	}
	if isDir && !stat.IsDir() {
		return fmt.Errorf("path %s is not a directory", pathAbs)
	}
	if !isDir && stat.IsDir() {
		return fmt.Errorf("path %s is a directory, must be a file", pathAbs)
	}
	return nil
}

func isCompressedFormat(ext string) bool {
	ext = strings.ToLower(ext)
	compressedExts := []string{
		".png", ".ico", ".jpg", ".jpeg", ".gif", ".webp",
		".zip", ".gz", ".br", ".zst",
		".woff", ".woff2",
	}

	for _, compExt := range compressedExts {
		if ext == compExt {
			return true
		}
	}
	return false
}

// filterZeroQuality removes encodings with q=0 or q=0.0 from Accept-Encoding header
func filterZeroQuality(acceptEncoding string) string {
	// If no q params, return unchanged (fast path for most requests)
	if !strings.Contains(acceptEncoding, "q=") {
		return acceptEncoding
	}

	var filtered []string
	parts := strings.Split(acceptEncoding, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Check if this encoding has q=0 or q=0.0
		normalized := strings.ReplaceAll(part, " ", "")
		if strings.HasSuffix(normalized, ";q=0") || strings.HasSuffix(normalized, ";q=0.0") {
			continue
		}

		filtered = append(filtered, part)
	}

	return strings.Join(filtered, ",")
}
