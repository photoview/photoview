package utils

import (
	"fmt"
	"log"
	"net/url"
	"strconv"
)

const apiPrefix = "/api"

func ApiListenUrl() *url.URL {
	const defaultIP = "127.0.0.1"
	const defaultPort = "4001"

	// Reuse the normalization logic so absolute/relative values are both supported.
	// If EnvAPIEndpoint contains a full URL, use its path component.
	apiPath := ApiEndpointUrl().Path
	if apiPath == "" {
		apiPath = apiPrefix
	}

	var listenAddr string

	listenAddr = EnvListenIP.GetValue()
	if listenAddr == "" {
		listenAddr = defaultIP
	}

	listenPortStr := EnvListenPort.GetValue()
	if listenPortStr == "" {
		listenPortStr = defaultPort
	}

	listenPort, err := strconv.Atoi(listenPortStr)
	if err != nil {
		log.Panicf("%q must be a number %q: %v", EnvListenPort.GetName(), listenPortStr, err)
	}

	apiURL, err := url.Parse(fmt.Sprintf("http://%s:%d", listenAddr, listenPort))
	if err != nil {
		log.Panicf("Could not format api url: %v", err)
	}
	apiURL.Path = apiPath

	return apiURL
}

func ApiEndpointUrl() *url.URL {
	apiEndpointStr := EnvAPIEndpoint.GetValue()
	if apiEndpointStr == "" {
		apiEndpointStr = apiPrefix
	}

	apiEndpointURL, err := url.Parse(apiEndpointStr)
	if err != nil {
		log.Panicf("ERROR: Environment variable %s is not a proper url (%s): %v",
			EnvAPIEndpoint.GetName(), EnvAPIEndpoint.GetValue(), err)
	}

	// If absolute URL with empty path (e.g. "https://host"), default to /api for backward compatibility.
	if apiEndpointURL.Host != "" && (apiEndpointURL.Path == "") {
		apiEndpointURL.Path = apiPrefix
	}
	// Ensure relative paths start with a leading slash.
	if apiEndpointURL.Host == "" &&
		apiEndpointURL.Scheme == "" &&
		apiEndpointURL.Path != "" &&
		apiEndpointURL.Path[0] != '/' {
		apiEndpointURL.Path = "/" + apiEndpointURL.Path
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
		log.Panicf("ERROR: Environment variable %s is not a proper url (%s): %v",
			EnvUIEndpoint.GetName(), EnvUIEndpoint.GetValue(), err)
	}

	return uiEndpointURL
}
