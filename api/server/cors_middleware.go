package server

import (
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/gorilla/mux"
)

func CORSMiddleware(devMode bool) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			methods := []string{http.MethodGet, http.MethodPost, http.MethodOptions}
			headers := []string{"authorization", "content-type"}

			w.Header().Set("Access-Control-Allow-Methods", strings.Join(methods, ","))
			w.Header().Set("Access-Control-Allow-Headers", strings.Join(headers, ","))

			endpoint, err := url.Parse(os.Getenv("API_ENDPOINT"))
			if err != nil {
				log.Fatalln("Could not parse API_ENDPOINT environment variable as url")
			}
			endpoint.Path = path.Join(endpoint.Path, "graphql")

			if devMode {
				// Development environment
				w.Header().Set("Access-Control-Allow-Origin", req.Header.Get("origin"))
				w.Header().Set("Vary", "Origin")
			} else {
				// Production environment
				publicEndpoint, err := url.Parse(os.Getenv("PUBLIC_ENDPOINT"))
				if err != nil {
					log.Printf("Error parsing environment variable PUBLIC_ENDPOINT as url: %s", err)
				} else {
					w.Header().Set("Access-Control-Allow-Origin", publicEndpoint.Scheme+"://"+publicEndpoint.Host)
				}
			}

			if req.Method != http.MethodOptions {
				next.ServeHTTP(w, req)
			} else {
				w.WriteHeader(200)
			}
		})
	}
}
