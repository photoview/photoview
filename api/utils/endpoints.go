package utils

import (
	"log"
	"net"
	"net/url"
	"strconv"
	"strings"
	"sync"
)

const defaultIP = "127.0.0.1"
const defaultPort = "4001"
const defaultAPIPrefix = "/api"

// Cached endpoint functions using sync.OnceValue for thread-safe lazy initialization
var (
	cachedApiEndpointUrl = sync.OnceValue(func() *url.URL {
		return computeApiEndpointUrl()
	})

	cachedApiListenUrl = sync.OnceValue(func() *url.URL {
		return computeApiListenUrl()
	})

	cachedUiEndpointUrls = sync.OnceValue(func() []*url.URL {
		return computeUiEndpointUrls()
	})
)

var (
	testApiEndpointUrl  *url.URL
	testApiListenUrl    *url.URL
	testUiEndpointUrls  []*url.URL
	testEndpointsLocker sync.RWMutex
)

// ConfigureTestEndpoints sets test-specific endpoint values.
// Used for testing to override cached values.
func ConfigureTestEndpoints(apiEndpoint, apiListen *url.URL, uiEndpoints []*url.URL) {
	testEndpointsLocker.Lock()
	defer testEndpointsLocker.Unlock()
	// Normalize apiEndpoint if provided
	if apiEndpoint != nil {
		normalized := *apiEndpoint
		normalized.Scheme = strings.ToLower(normalized.Scheme)
		normalized.Host = strings.ToLower(normalized.Host)
		testApiEndpointUrl = &normalized
	} else {
		testApiEndpointUrl = nil
	}

	// Normalize apiListen if provided
	if apiListen != nil {
		normalized := *apiListen
		normalized.Scheme = strings.ToLower(normalized.Scheme)
		normalized.Host = strings.ToLower(normalized.Host)
		testApiListenUrl = &normalized
	} else {
		testApiListenUrl = nil
	}

	// Normalize uiEndpoints if provided
	if uiEndpoints != nil {
		testUiEndpointUrls = make([]*url.URL, 0, len(uiEndpoints))
		for _, endpoint := range uiEndpoints {
			if endpoint != nil {
				normalized := *endpoint
				normalized.Scheme = strings.ToLower(normalized.Scheme)
				normalized.Host = strings.ToLower(normalized.Host)
				testUiEndpointUrls = append(testUiEndpointUrls, &normalized)
			}
		}
	} else {
		testUiEndpointUrls = nil
	}
}

// ResetTestEndpoints clears all test endpoint overrides.
// Call this in t.Cleanup() or defer to ensure test isolation.
func ResetTestEndpoints() {
	testEndpointsLocker.Lock()
	defer testEndpointsLocker.Unlock()
	testApiEndpointUrl = nil
	testApiListenUrl = nil
	testUiEndpointUrls = nil
}

func ApiListenUrl() *url.URL {
	testEndpointsLocker.RLock()
	testUrl := testApiListenUrl
	testEndpointsLocker.RUnlock()
	if testUrl != nil {
		urlCopy := *testUrl
		return &urlCopy
	}

	cached := cachedApiListenUrl()
	urlCopy := *cached
	return &urlCopy
}

// ComputeApiListenUrlForTest exposes computeApiListenUrl for testing.
// This allows tests to verify validation logic without being affected by caching.
func ComputeApiListenUrlForTest() *url.URL {
	return computeApiListenUrl()
}

func computeApiListenUrl() *url.URL {
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
	testEndpointsLocker.RLock()
	testUrl := testApiEndpointUrl
	testEndpointsLocker.RUnlock()
	if testUrl != nil {
		urlCopy := *testUrl
		return &urlCopy
	}

	cached := cachedApiEndpointUrl()
	urlCopy := *cached
	return &urlCopy
}

// ComputeApiEndpointUrlForTest exposes computeApiEndpointUrl for testing.
// This allows tests to verify validation logic without being affected by caching.
func ComputeApiEndpointUrlForTest() *url.URL {
	return computeApiEndpointUrl()
}

func computeApiEndpointUrl() *url.URL {
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

// UiEndpointUrls returns a list of allowed UI endpoints.
// Returns nil if UI is served by this server (no external UI).
func UiEndpointUrls() []*url.URL {
	testEndpointsLocker.RLock()
	testUrls := testUiEndpointUrls
	testEndpointsLocker.RUnlock()
	if testUrls != nil {
		copies := make([]*url.URL, len(testUrls))
		for i, u := range testUrls {
			urlCopy := *u
			copies[i] = &urlCopy
		}
		return copies
	}

	cached := cachedUiEndpointUrls()
	if cached == nil {
		return nil
	}

	// Return copies of the URLs to prevent mutation
	copies := make([]*url.URL, len(cached))
	for i, u := range cached {
		urlCopy := *u
		copies[i] = &urlCopy
	}
	return copies
}

// ComputeUiEndpointUrlsForTest exposes computeUiEndpointUrls for testing.
// This allows tests to verify validation and port normalization logic without being affected by caching.
func ComputeUiEndpointUrlsForTest() []*url.URL {
	return computeUiEndpointUrls()
}

func computeUiEndpointUrls() []*url.URL {
	shouldServeUI := ShouldServeUI()
	if shouldServeUI {
		return nil
	}

	endpointStr := strings.ToLower(EnvUIEndpoints.GetValue())
	if endpointStr == "" {
		log.Panic("ERROR: PHOTOVIEW_UI_ENDPOINTS must be set when PHOTOVIEW_SERVE_UI=0, but is empty or unset")
	}

	// Split by comma and trim whitespace
	endpointStrings := strings.Split(endpointStr, ",")
	endpoints := make([]*url.URL, 0, len(endpointStrings))

	for _, urlStr := range endpointStrings {
		urlStr = strings.TrimSpace(urlStr)
		if urlStr == "" {
			continue
		}

		parsedURL, err := url.Parse(urlStr)
		if err != nil {
			log.Printf("ERROR: Invalid URL in %s (%s): %v\n", EnvUIEndpoints.GetName(), urlStr, err)
			continue
		}

		// Validate required components
		if parsedURL.Scheme == "" || parsedURL.Host == "" {
			log.Printf("ERROR: UI endpoint URL must include scheme and host: %s\n", urlStr)
			continue
		}

		var skip bool
		endpoints, skip = managePort(parsedURL, endpoints, urlStr)
		if skip {
			continue
		}

		if !contains(endpoints, parsedURL) {
			endpoints = append(endpoints, parsedURL)
		}
	}

	if len(endpoints) == 0 {
		log.Panicf("ERROR: No valid UI endpoints found in %s", EnvUIEndpoints.GetName())
	}

	return endpoints
}

func managePort(parsedURL *url.URL, endpoints []*url.URL, urlStr string) ([]*url.URL, bool) {
	// Add default ports for http/https if missing, or add host without port if standard port is specified
	host := parsedURL.Hostname()
	port := parsedURL.Port()
	scheme := parsedURL.Scheme

	switch scheme {
	case "http":
		switch port {
		case "":
			// Add with default port 80
			extendedURL, _ := url.Parse("http://" + host + ":80")
			if !contains(endpoints, extendedURL) {
				endpoints = append(endpoints, extendedURL)
			}
		case "80":
			// Add without port
			simpleURL, _ := url.Parse("http://" + host)
			if !contains(endpoints, simpleURL) {
				endpoints = append(endpoints, simpleURL)
			}
		}
	case "https":
		switch port {
		case "":
			// Add with default port 443
			extendedURL, _ := url.Parse("https://" + host + ":443")
			if !contains(endpoints, extendedURL) {
				endpoints = append(endpoints, extendedURL)
			}
		case "443":
			// Add without port
			simpleURL, _ := url.Parse("https://" + host)
			if !contains(endpoints, simpleURL) {
				endpoints = append(endpoints, simpleURL)
			}
		}
	default:
		log.Printf("ERROR: Unknown scheme in UI endpoint URL (must be http or https): %s\n", urlStr)
		return endpoints, true
	}
	return endpoints, false
}

// contains checks if the given *url.URL is present in the slice.
func contains(urls []*url.URL, target *url.URL) bool {
	for _, u := range urls {
		if u.String() == target.String() {
			return true
		}
	}
	return false
}
