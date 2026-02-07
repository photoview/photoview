package server

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/mux"
	"github.com/photoview/photoview/api/utils"
)

// Variable to allow mocking in tests
var uiEndpointUrlsFunc = func() []*url.URL {
	return utils.UiEndpointUrls()
}

func CORSMiddleware(devMode bool) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

			var allowedOrigin string

			if devMode {
				// Development environment
				w.Header().Set("Access-Control-Allow-Origin", req.Header.Get("Origin"))
				w.Header().Set("Vary", "Origin")
			} else {
				// Production environment
				uiEndpoints := uiEndpointUrlsFunc()
				allowedOrigin = setAllowedCORSOrigin(uiEndpoints, req, w)
			}

			corsEnabled := devMode || allowedOrigin != ""
			w = handleCORS(corsEnabled, w)

			if req.Method != http.MethodOptions {
				next.ServeHTTP(w, req)
			} else {
				w.WriteHeader(200)
			}
		})
	}
}

// setAllowedCORSOrigin checks if the request's Origin header matches any of the allowed UI endpoint URLs.
// If a match is found, it sets the appropriate CORS headers on the response and returns the matched origin string.
// If no match is found, it returns an empty string.
func setAllowedCORSOrigin(uiEndpoints []*url.URL, req *http.Request, w http.ResponseWriter) string {
	var matchedOrigin string
	if uiEndpoints != nil && len(uiEndpoints) > 0 {
		requestOrigin := req.Header.Get("Origin")
		if requestOrigin != "" {
			// Check if request origin matches any allowed endpoint
			if matchedOrigin = findMatchingOrigin(requestOrigin, uiEndpoints); matchedOrigin != "" {
				w.Header().Set("Access-Control-Allow-Origin", matchedOrigin)
				w.Header().Set("Vary", "Origin")
			}
		}
	}
	return matchedOrigin
}

func findMatchingOrigin(requestOrigin string, allowedEndpoints []*url.URL) string {
	requestURL, err := url.Parse(requestOrigin)
	if err != nil {
		return ""
	}

	requestOriginStr := strings.ToLower(requestURL.Scheme + "://" + requestURL.Host)

	for _, endpoint := range allowedEndpoints {
		allowedOriginStr := endpoint.Scheme + "://" + endpoint.Host
		if requestOriginStr == allowedOriginStr {
			return requestURL.Scheme + "://" + requestURL.Host
		}
	}

	return ""
}

func handleCORS(corsEnabled bool, w http.ResponseWriter) http.ResponseWriter {
	if corsEnabled {
		methods := []string{http.MethodGet, http.MethodPost, http.MethodOptions}
		requestHeaders := []string{"authorization", "content-type", "content-length", "TokenPassword"}
		responseHeaders := []string{"content-length"}

		w.Header().Set("Access-Control-Allow-Methods", strings.Join(methods, ", "))
		w.Header().Set("Access-Control-Allow-Headers", strings.Join(requestHeaders, ", "))
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Expose-Headers", strings.Join(responseHeaders, ", "))
	}
	return w
}
