package server

import (
	"net/http"
	"path"
	"strings"

	"github.com/gorilla/mux"
	"github.com/viktorstrate/photoview/api/utils"
)

func CORSMiddleware(devMode bool) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

			methods := []string{http.MethodGet, http.MethodPost, http.MethodOptions}
			headers := []string{"authorization", "content-type", "content-length", "TokenPassword"}

			w.Header().Set("Access-Control-Allow-Methods", strings.Join(methods, ","))
			w.Header().Set("Access-Control-Allow-Headers", strings.Join(headers, ","))
			w.Header().Set("Access-Control-Expose-Headers", "content-length")

			endpoint := utils.ApiEndpointUrl()
			endpoint.Path = path.Join(endpoint.Path, "graphql")

			if devMode {
				// Development environment
				w.Header().Set("Access-Control-Allow-Origin", req.Header.Get("origin"))
				w.Header().Set("Vary", "Origin")
			} else {
				// Production environment
				uiEndpoint := utils.UiEndpointUrl()
				w.Header().Set("Access-Control-Allow-Origin", uiEndpoint.Scheme+"://"+uiEndpoint.Host)
			}

			if req.Method != http.MethodOptions {
				next.ServeHTTP(w, req)
			} else {
				w.WriteHeader(200)
			}
		})
	}
}
