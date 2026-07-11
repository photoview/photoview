package server

import (
	"errors"
	"log"
	"net/http"
	"net/url"

	"github.com/99designs/gqlgen/graphql/handler/transport"
	coderws "github.com/coder/websocket"
	"github.com/photoview/photoview/api/utils"
)

type websocketImplementation struct {
	accept      transport.CoderWebsocketImplementation
	checkOrigin func(*http.Request) bool
}

func WebsocketImplementation(devMode bool) transport.WebsocketImplementation {
	return websocketImplementation{
		accept: transport.CoderWebsocketImplementation{
			AcceptOptions: coderws.AcceptOptions{
				// We keep the project's existing origin policy in checkOrigin
				// below, so gqlgen's adapter should not apply a second,
				// different policy on top of it.
				InsecureSkipVerify: true,
			},
		},
		checkOrigin: websocketCheckOrigin(devMode),
	}
}

func (w websocketImplementation) Accept(rw http.ResponseWriter, r *http.Request, options transport.WebsocketAcceptOptions) (transport.WebsocketConn, error) {
	if !w.checkOrigin(r) {
		return nil, errors.New("websocket origin not allowed")
	}

	return w.accept.Accept(rw, r, options)
}

func websocketCheckOrigin(devMode bool) func(r *http.Request) bool {
	return func(r *http.Request) bool {
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

			return isUIOnSameHost(uiEndpoint, originURL)
		}
	}
}

func isUIOnSameHost(uiEndpoint *url.URL, originURL *url.URL) bool {
	if uiEndpoint.Host == originURL.Host {
		return true
	} else {
		log.Printf("Not allowing websocket request from %s because it doesn't match PHOTOVIEW_UI_ENDPOINT %s",
			originURL.Host, uiEndpoint.Host)
		return false
	}
}
