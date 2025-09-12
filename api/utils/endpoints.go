package utils

import (
	"log"
	"net"
	"net/url"
	"strconv"
	"strings"
)

const defaultIP = "127.0.0.1"
const defaultPort = "4001"
const defaultAPIPrefix = "/api"

func ApiListenUrl() *url.URL {
	apiPath := ApiEndpointUrl().Path

	var listenAddr string

	listenAddr = EnvListenIP.GetValue()
	if listenAddr == "" {
		listenAddr = defaultIP
	}
	// Validate listenAddr is a valid IP (IPv4 or IPv6)
	if parsedIP := net.ParseIP(listenAddr); parsedIP == nil {
		log.Panicf("%q must be a valid IPv4 or IPv6 address: %q", EnvListenIP.GetName(), listenAddr)
	}

	listenPortStr := EnvListenPort.GetValue()
	if listenPortStr == "" {
		listenPortStr = defaultPort
	}
	listenPort, err := strconv.Atoi(listenPortStr)
	if err != nil {
		log.Panicf("%q must be a number %q: %v", EnvListenPort.GetName(), listenPortStr, err)
	}
	if listenPort < 1 || listenPort > 65535 {
		log.Panicf("%q must be a valid port number (1-65535): %d", EnvListenPort.GetName(), listenPort)
	}

	return &url.URL{
		Scheme: "http",
		Host:   net.JoinHostPort(listenAddr, listenPortStr),
		Path:   apiPath,
	}
}

func ApiEndpointUrl() *url.URL {
	apiEndpointStr := EnvAPIEndpoint.GetValue()
	if apiEndpointStr == "" {
		apiEndpointStr = defaultAPIPrefix
	}

	apiEndpointURL, err := url.Parse(apiEndpointStr)
	if err != nil {
		log.Panicf("ERROR: Environment variable %s is not a proper url (%s): %v",
			EnvAPIEndpoint.GetName(), EnvAPIEndpoint.GetValue(), err)
	}

	// If absolute URL with empty path (e.g. "https://host"), default to /api for backward compatibility.
	if apiEndpointURL.Scheme != "" && apiEndpointURL.Host != "" && apiEndpointURL.Path == "" {
		apiEndpointURL.Path = defaultAPIPrefix
	}
	// Ensure relative paths start with a leading slash.
	if apiEndpointURL.Scheme == "" &&
		apiEndpointURL.Host == "" &&
		apiEndpointURL.Path != "" &&
		!strings.HasPrefix(apiEndpointURL.Path, "/") {
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
