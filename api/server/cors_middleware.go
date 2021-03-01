package server

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/mux"
	"github.com/photoview/photoview/api/utils"
)

func CORSMiddleware(devMode bool) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

			var uiEndpoint *url.URL = nil

			if devMode {
				// Development environment
				w.Header().Set("Access-Control-Allow-Origin", req.Header.Get("origin"))
				w.Header().Set("Vary", "Origin")
			} else {
				// Production environment
				uiEndpoint = utils.UiEndpointUrl()
				if uiEndpoint != nil {
					// Only allow CORS if UI endpoint is defined
					w.Header().Set("Access-Control-Allow-Origin", uiEndpoint.Scheme+"://"+uiEndpoint.Host)
				}
			}

			corsEnabled := devMode || uiEndpoint != nil
			if corsEnabled {
				methods := []string{http.MethodGet, http.MethodPost, http.MethodOptions}
				requestHeaders := []string{"authorization", "content-type", "content-length", "TokenPassword"}
				responseHeaders := []string{"content-length"}

				w.Header().Set("Access-Control-Allow-Methods", strings.Join(methods, ", "))
				w.Header().Set("Access-Control-Allow-Headers", strings.Join(requestHeaders, ", "))
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Access-Control-Expose-Headers", strings.Join(responseHeaders, ", "))
			}

			if req.Method != http.MethodOptions {
				next.ServeHTTP(w, req)
			} else {
				w.WriteHeader(200)
			}
		})
	}
}
