package utils

import (
	"fmt"
	"log"
	"net/url"
	"path"
	"strconv"
)

func ApiListenUrl() *url.URL {
	const defaultPort = "4001"

	shouldServeUI := ShouldServeUI()

	apiPrefix := "/"
	if shouldServeUI {
		apiPrefix = "/api"
	}

	var listenAddr string

	listenAddr = EnvListenIP.GetValue()
	if listenAddr == "" {
		listenAddr = "127.0.0.1"
	}

	listenPortStr := EnvListenPort.GetValue()
	if listenPortStr == "" {
		listenPortStr = defaultPort
	}

	listenPort, err := strconv.Atoi(listenPortStr)
	if err != nil {
		log.Fatalf("%s must be a number: '%s'\n%s", EnvListenPort.GetName(), listenPortStr, err)
	}

	apiUrl, err := url.Parse(fmt.Sprintf("http://%s:%d", listenAddr, listenPort))
	if err != nil {
		log.Fatalf("Could not format api url: %s", err)
	}
	apiUrl.Path = apiPrefix

	return apiUrl
}

func ApiEndpointUrl() *url.URL {
	var apiEndpointStr string

	shouldServeUI := ShouldServeUI()
	if shouldServeUI {
		apiEndpointStr = "/"
	} else {
		apiEndpointStr = EnvAPIEndpoint.GetValue()
	}

	apiEndpointURL, err := url.Parse(apiEndpointStr)
	if err != nil {
		log.Fatalf("ERROR: Environment variable %s is not a proper url (%s)", EnvAPIEndpoint.GetName(), EnvAPIEndpoint.GetValue())
	}

	if shouldServeUI {
		apiEndpointURL.Path = path.Join(apiEndpointURL.Path, "/api")
	}

	return apiEndpointURL
}

func UiEndpointUrl() *url.URL {
	shouldServeUI := ShouldServeUI()
	if shouldServeUI {
		return nil
	}

	uiEndpointURL, err := url.Parse(EnvUIEndpoint.GetValue())
	if err != nil {
		log.Fatalf("ERROR: Environment variable %s is not a proper url (%s)", EnvUIEndpoint.GetName(), EnvUIEndpoint.GetValue())
	}

	return uiEndpointURL
}
