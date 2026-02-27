package server

import (
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/photoview/photoview/api/utils"
)

func WebsocketUpgrader(devMode bool) websocket.Upgrader {
	return websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			if devMode {
				return true
			}

			if utils.ShouldServeUI() {
				return true
			}

			originHeader := r.Header.Get("Origin")
			if originHeader == "" {
				return true
			}

			originURL, err := url.Parse(originHeader)
			if err != nil {
				log.Printf("Could not parse origin header of websocket request: %s", err)
				return false
			}

			uiEndpoints := utils.UiEndpointUrls()
			for _, uiEndpoint := range uiEndpoints {
				if uiEndpoint.Scheme+uiEndpoint.Host == strings.ToLower(originURL.Scheme+originURL.Host) {
					return true
				}
			}

			// Log rejection with sanitization
			sanitizedOriginHost := strings.ReplaceAll(originURL.Host, "\n", "\\n")
			sanitizedOriginHost = strings.ReplaceAll(sanitizedOriginHost, "\r", "\\r")
			allowedHosts := make([]string, len(uiEndpoints))
			for i, ep := range uiEndpoints {
				allowedHosts[i] = ep.Host
			}
			log.Printf(
				"Rejected websocket request from %s because it doesn't match allowed hosts in the PHOTOVIEW_UI_ENDPOINTS: %v",
				sanitizedOriginHost, allowedHosts)
			return false
		},
	}
}
