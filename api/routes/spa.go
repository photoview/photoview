package routes

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/photoview/photoview/api/utils"
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
	if r.URL.Path == "/" || r.URL.Path == "/index.html" || os.IsNotExist(err) {
		// serve index.html, if index.html is asked for, or if file does not exist
		bytes, err := h.renderIndexHtml()
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to renderIndexHtml as %v", err.Error()), http.StatusInternalServerError)
			return
		}

		w.Write(bytes)
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

func (h SpaHandler) renderIndexHtml() ([]byte, error) {
	indexFilePath := filepath.Join(h.staticPath, h.indexPath)

	f, err := os.Open(indexFilePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	bytes, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	if utils.GuestAccepted() {
		return []byte(strings.ReplaceAll(string(bytes), "<head>", "<head><script>window.GuestAccepted=true;</script>")), nil
	} else {
		return bytes, nil
	}

}
