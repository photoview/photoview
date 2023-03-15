package routes

import (
	"net/http"
	"os"
	"path"
	"path/filepath"
)

// SpaHandler implements the http.Handler interface, so we can use it
// to respond to HTTP requests. The path to the static directory and
// path to the index file within that static directory are used to
// serve the SPA in the given static directory.
type SpaHandler struct {
	staticPath string
	indexPath  string
}

func NewSpaHandler(staticPath string, indexPath string) SpaHandler {
	return SpaHandler{
		indexPath:  indexPath,
		staticPath: staticPath,
	}
}

// ServeHTTP inspects the URL path to locate a file within the static dir
// on the SPA handler. If a file is found, it will be served. If not, the
// file located at the index path on the SPA handler will be served. This
// is suitable behavior for serving an SPA (single page application).
func (h SpaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// get the clean path to prevent directory traversal
	servePath := path.Clean(r.URL.Path)

	// prepend the path with the path to the static directory
	servePath = filepath.Join(h.staticPath, servePath)

	// check whether a file exists at the given path
	_, err := os.Stat(servePath)
	if os.IsNotExist(err) {
		// file does not exist, serve index.html
		http.ServeFile(w, r, filepath.Join(h.staticPath, h.indexPath))
		return
	} else if err != nil {
		// if we got an error (that wasn't that the file doesn't exist) stating the
		// file, return a 500 internal server error and stop
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// otherwise, use http.FileServer to serve the static dir
	http.FileServer(http.Dir(h.staticPath)).ServeHTTP(w, r)
}
