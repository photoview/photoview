package server

import (
	"log"
	"net/http"
	"net/url"

	"github.com/gorilla/websocket"
	"github.com/photoview/photoview/api/utils"
)

func WebsocketUpgrader(devMode bool) websocket.Upgrader {
	return websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			if devMode {
				return true
			} else {
				uiEndpoint := utils.UiEndpointUrl()
				if uiEndpoint == nil {
					return true
				}

				if r.Header.Get("origin") == "" {
					return true
				}

				originURL, err := url.Parse(r.Header.Get("origin"))
				if err != nil {
					log.Printf("Could not parse origin header of websocket request: %s", err)
					return false
				}

				if uiEndpoint.Host == originURL.Host {
					return true
				} else {
					log.Printf("Not allowing websocket request from %s because it doesn't match PHOTOVIEW_UI_ENDPOINT %s", originURL.Host, uiEndpoint.Host)
					return false
				}
			}
		},
	}
}
