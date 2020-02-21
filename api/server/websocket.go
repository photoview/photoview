package server

import (
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/gorilla/websocket"
)

func WebsocketUpgrader(devMode bool) websocket.Upgrader {
	return websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			if devMode {
				return true
			} else {
				pubEndpoint, err := url.Parse(os.Getenv("PUBLIC_ENDPOINT"))
				if err != nil {
					log.Printf("Could not parse API_ENDPOINT environment variable as url: %s", err)
					return false
				}

				if r.Header.Get("origin") == "" {
					return true
				}

				originURL, err := url.Parse(r.Header.Get("origin"))
				if err != nil {
					log.Printf("Could not parse origin header of websocket request: %s", err)
					return false
				}

				if pubEndpoint.Host == originURL.Host {
					return true
				} else {
					log.Printf("Not allowing websocket request from %s because it doesn't match PUBLIC_ENDPOINT %s", originURL.Host, pubEndpoint.Host)
					return false
				}
			}
		},
	}
}
