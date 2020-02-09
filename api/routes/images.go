package routes

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
)

func ImageRoutes() chi.Router {
	router := chi.NewRouter()
	router.Get("/{name}", func(w http.ResponseWriter, r *http.Request) {
		image_name := chi.URLParam(r, "name")
		w.Write([]byte(fmt.Sprintf("Image: %s", image_name)))
	})

	return router
}
