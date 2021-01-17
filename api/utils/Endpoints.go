package utils

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"path"
	"strconv"
)

func ApiListenUrl() *url.URL {
	const defaultPort = "4001"

	shouldServeUI := os.Getenv("PHOTOVIEW_SERVE_UI") == "1"

	apiPrefix := "/"
	if shouldServeUI {
		apiPrefix = "/api"
	}

	var listenAddr string

	listenAddr = os.Getenv("PHOTOVIEW_LISTEN_IP")
	if listenAddr == "" {
		listenAddr = "127.0.0.1"
	}

	listenPortStr := os.Getenv("PHOTOVIEW_LISTEN_PORT")
	if listenPortStr == "" {
		listenPortStr = defaultPort
	}

	listenPort, err := strconv.Atoi(listenPortStr)
	if err != nil {
		log.Fatalf("PHOTOVIEW_LISTEN_PORT must be a number: '%s'\n%s", listenPortStr, err)
	}

	apiUrl, err := url.Parse(fmt.Sprintf("http://%s:%d", listenAddr, listenPort))
	if err != nil {
		log.Fatalf("Could not format api url: %s", err)
	}
	apiUrl.Path = apiPrefix

	return apiUrl
}

func ApiEndpointUrl() *url.URL {
	apiEndpointStr := os.Getenv("PHOTOVIEW_API_ENDPOINT")

	shouldServeUI := os.Getenv("PHOTOVIEW_SERVE_UI") == "1"
	if shouldServeUI {
		apiEndpointStr = os.Getenv("PHOTOVIEW_PUBLIC_ENDPOINT")
	}

	apiEndpointUrl, err := url.Parse(apiEndpointStr)
	if err != nil {
		log.Fatalf("ERROR: Environment variable PHOTOVIEW_API_ENDPOINT is not a proper url")
	}

	if shouldServeUI {
		apiEndpointUrl.Path = path.Join(apiEndpointUrl.Path, "/api")
	}

	return apiEndpointUrl
}

func UiEndpointUrl() *url.URL {
	uiEndpointStr := os.Getenv("PHOTOVIEW_UI_ENDPOINT")

	shouldServeUI := os.Getenv("PHOTOVIEW_SERVE_UI") == "1"
	if shouldServeUI {
		uiEndpointStr = os.Getenv("PHOTOVIEW_PUBLIC_ENDPOINT")
	}

	uiEndpointUrl, err := url.Parse(uiEndpointStr)
	if err != nil {
		log.Fatalf("ERROR: Environment variable PHOTOVIEW_UI_ENDPOINT is not a proper url")
	}

	return uiEndpointUrl
}
