package server

import (
	"errors"
	"log"
	"net/http"
	"net/url"
	"strings"

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
	}
}
