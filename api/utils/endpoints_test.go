package utils_test

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/photoview/photoview/api/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Test Helper Functions
// =============================================================================

// mustParseURL is a helper for constructing test URLs
func mustParseURL(rawURL string) *url.URL {
	u, err := url.Parse(rawURL)
	if err != nil {
		panic(fmt.Sprintf("Invalid test URL %q: %v", rawURL, err))
	}
	return u
}

// =============================================================================
// ApiListenUrl Tests - Using ConfigureTestEndpoints for accessor behavior
// =============================================================================

func TestApiListenUrl(t *testing.T) {
	tests := []struct {
		name       string
		listenURL  *url.URL
		wantScheme string
		wantHost   string
		wantPath   string
	}{
		{
			name:       "default values",
			listenURL:  mustParseURL("http://127.0.0.1:4001/api"),
			wantScheme: "http",
			wantHost:   "127.0.0.1:4001",
			wantPath:   "/api",
		},
		{
			name:       "custom IP and port",
			listenURL:  mustParseURL("http://192.168.1.100:8080/api"),
			wantScheme: "http",
			wantHost:   "192.168.1.100:8080",
			wantPath:   "/api",
		},
		{
			name:       "IPv6 address",
			listenURL:  mustParseURL("http://[::1]:4001/api"),
			wantScheme: "http",
			wantHost:   "[::1]:4001",
			wantPath:   "/api",
		},
		{
			name:       "custom API path",
			listenURL:  mustParseURL("http://127.0.0.1:4001/custom"),
			wantScheme: "http",
			wantHost:   "127.0.0.1:4001",
			wantPath:   "/custom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			utils.ConfigureTestEndpoints(nil, tt.listenURL, nil)
			t.Cleanup(utils.ResetTestEndpoints)

			url := utils.ApiListenUrl()
			require.NotNil(t, url, "ApiListenUrl should return a URL")
			assert.Equal(t, tt.wantScheme, url.Scheme, "Scheme mismatch")
			assert.Equal(t, tt.wantHost, url.Host, "Host mismatch")
			assert.Equal(t, tt.wantPath, url.Path, "Path mismatch")
		})
	}
}

func TestApiListenUrlValidation(t *testing.T) {
	tests := []struct {
		name       string
		listenIP   string
		listenPort string
		wantPanic  bool
	}{
		{
			name:       "invalid IP",
			listenIP:   "invalid-ip",
			listenPort: "4001",
			wantPanic:  true,
		},
		{
			name:       "invalid port - non-numeric",
			listenIP:   "127.0.0.1",
			listenPort: "not-a-number",
			wantPanic:  true,
		},
		{
			name:       "invalid port - too low",
			listenIP:   "127.0.0.1",
			listenPort: "0",
			wantPanic:  true,
		},
		{
			name:       "invalid port - too high",
			listenIP:   "127.0.0.1",
			listenPort: "65536",
			wantPanic:  true,
		},
		{
			name:       "valid IP and port",
			listenIP:   "127.0.0.1",
			listenPort: "4001",
			wantPanic:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("PHOTOVIEW_LISTEN_IP", tt.listenIP)
			t.Setenv("PHOTOVIEW_LISTEN_PORT", tt.listenPort)
			t.Setenv("PHOTOVIEW_API_ENDPOINT", "/api")

			if tt.wantPanic {
				assert.Panics(t, func() {
					utils.ComputeApiListenUrlForTest()
				}, "Expected panic for invalid input")
			} else {
				assert.NotPanics(t, func() {
					url := utils.ComputeApiListenUrlForTest()
					assert.NotNil(t, url)
				}, "Should not panic for valid input")
			}
		})
	}
}

func TestApiListenUrlReturnsCopy(t *testing.T) {
	testURL := mustParseURL("http://127.0.0.1:4001/api")
	utils.ConfigureTestEndpoints(nil, testURL, nil)
	t.Cleanup(utils.ResetTestEndpoints)

	url1 := utils.ApiListenUrl()
	url2 := utils.ApiListenUrl()

	require.NotNil(t, url1)
	require.NotNil(t, url2)
	assert.NotSame(t, url1, url2, "Should return different URL instances")

	url1.Path = "/modified"
	assert.Equal(t, "/api", url2.Path, "Mutation of one URL should not affect another")

	url3 := utils.ApiListenUrl()
	assert.Equal(t, "/api", url3.Path, "Subsequent calls should return unmodified path")
}

// =============================================================================
// ApiEndpointUrl Tests
// =============================================================================

func TestApiEndpointUrl(t *testing.T) {
	tests := []struct {
		name       string
		endpoint   *url.URL
		wantScheme string
		wantHost   string
		wantPath   string
	}{
		{
			name:       "default value",
			endpoint:   mustParseURL("/api"),
			wantScheme: "",
			wantHost:   "",
			wantPath:   "/api",
		},
		{
			name:       "custom relative path",
			endpoint:   mustParseURL("/custom/api"),
			wantScheme: "",
			wantHost:   "",
			wantPath:   "/custom/api",
		},
		{
			name:       "absolute URL with path",
			endpoint:   mustParseURL("https://example.com/api"),
			wantScheme: "https",
			wantHost:   "example.com",
			wantPath:   "/api",
		},
		{
			name:       "absolute URL with port",
			endpoint:   mustParseURL("http://example.com:8080/api"),
			wantScheme: "http",
			wantHost:   "example.com:8080",
			wantPath:   "/api",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			utils.ConfigureTestEndpoints(tt.endpoint, nil, nil)
			t.Cleanup(utils.ResetTestEndpoints)

			url := utils.ApiEndpointUrl()
			require.NotNil(t, url, "ApiEndpointUrl should return a URL")
			assert.Equal(t, tt.wantScheme, url.Scheme, "Scheme mismatch")
			assert.Equal(t, tt.wantHost, url.Host, "Host mismatch")
			assert.Equal(t, tt.wantPath, url.Path, "Path mismatch")
		})
	}
}

func TestApiEndpointUrlReturnsCopy(t *testing.T) {
	testEndpoint := mustParseURL("/api")
	utils.ConfigureTestEndpoints(testEndpoint, nil, nil)
	t.Cleanup(utils.ResetTestEndpoints)

	url1 := utils.ApiEndpointUrl()
	url2 := utils.ApiEndpointUrl()

	require.NotNil(t, url1)
	require.NotNil(t, url2)
	assert.NotSame(t, url1, url2, "Should return different URL instances")

	url1.Path = "/modified"
	assert.Equal(t, "/api", url2.Path, "Mutation of one URL should not affect another")

	url3 := utils.ApiEndpointUrl()
	assert.Equal(t, "/api", url3.Path, "Subsequent calls should return unmodified path")
}

// =============================================================================
// UiEndpointUrls Tests
// =============================================================================

func TestUiEndpointUrls(t *testing.T) {
	tests := []struct {
		name      string
		endpoints []*url.URL
		wantCount int
		validate  func(t *testing.T, urls []*url.URL)
	}{
		{
			name: "single endpoint",
			endpoints: []*url.URL{
				mustParseURL("https://ui.example.com"),
			},
			wantCount: 1,
			validate: func(t *testing.T, urls []*url.URL) {
				assert.Equal(t, "https", urls[0].Scheme)
				assert.Equal(t, "ui.example.com", urls[0].Host)
			},
		},
		{
			name: "multiple endpoints",
			endpoints: []*url.URL{
				mustParseURL("https://ui1.example.com"),
				mustParseURL("https://ui2.example.com"),
				mustParseURL("http://localhost:3000"),
			},
			wantCount: 3,
			validate: func(t *testing.T, urls []*url.URL) {
				assert.Equal(t, "ui1.example.com", urls[0].Host)
				assert.Equal(t, "ui2.example.com", urls[1].Host)
				assert.Equal(t, "localhost:3000", urls[2].Host)
			},
		},
		{
			name: "endpoints with paths",
			endpoints: []*url.URL{
				mustParseURL("https://ui.example.com/app"),
				mustParseURL("https://ui2.example.com/photoview"),
			},
			wantCount: 2,
			validate: func(t *testing.T, urls []*url.URL) {
				assert.Equal(t, "/app", urls[0].Path)
				assert.Equal(t, "/photoview", urls[1].Path)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			utils.ConfigureTestEndpoints(nil, nil, tt.endpoints)
			t.Cleanup(utils.ResetTestEndpoints)

			urls := utils.UiEndpointUrls()

			require.NotNil(t, urls)
			require.Len(t, urls, tt.wantCount)

			if tt.validate != nil {
				tt.validate(t, urls)
			}
		})
	}
}

func TestUiEndpointUrlsValidation(t *testing.T) {
	tests := []struct {
		name      string
		serveUI   string
		endpoints string
		wantPanic bool
	}{
		{
			name:      "empty endpoints when not serving UI",
			serveUI:   "0",
			endpoints: "",
			wantPanic: true,
		},
		{
			name:      "no valid URLs panics",
			serveUI:   "0",
			endpoints: "invalid,/just/path",
			wantPanic: true,
		},
		{
			name:      "valid URL does not panic",
			serveUI:   "0",
			endpoints: "https://valid.com",
			wantPanic: false,
		},
		{
			name:      "serving UI returns nil without panic",
			serveUI:   "1",
			endpoints: "",
			wantPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("PHOTOVIEW_SERVE_UI", tt.serveUI)
			t.Setenv("PHOTOVIEW_UI_ENDPOINTS", tt.endpoints)

			if tt.wantPanic {
				assert.Panics(t, func() {
					utils.ComputeUiEndpointUrlsForTest()
				}, "Expected panic for invalid configuration")
			} else {
				assert.NotPanics(t, func() {
					urls := utils.ComputeUiEndpointUrlsForTest()
					if tt.serveUI == "1" {
						assert.Nil(t, urls, "Should return nil when serving UI")
					} else {
						assert.NotNil(t, urls, "Should return URLs when not serving UI")
					}
				}, "Should not panic for valid configuration")
			}
		})
	}
}

func TestUiEndpointUrlsReturnsCopies(t *testing.T) {
	testEndpoints := []*url.URL{
		mustParseURL("https://ui.example.com"),
	}
	utils.ConfigureTestEndpoints(nil, nil, testEndpoints)
	t.Cleanup(utils.ResetTestEndpoints)

	urls1 := utils.UiEndpointUrls()
	urls2 := utils.UiEndpointUrls()

	require.NotNil(t, urls1)
	require.NotNil(t, urls2)
	require.Len(t, urls1, 1)
	require.Len(t, urls2, 1)

	// Verify URL pointers are different (preventing mutation issues)
	assert.NotSame(t, urls1[0], urls2[0], "Should return different URL instances")

	// Mutate and verify isolation
	urls1[0].Path = "/modified"
	assert.Empty(t, urls2[0].Path, "Mutation of one URL should not affect another copy")

	urls3 := utils.UiEndpointUrls()
	require.Len(t, urls3, 1)
	assert.Empty(t, urls3[0].Path, "Subsequent calls should return unmodified URLs")
}

// =============================================================================
// Port Normalization Tests - Testing production computation logic
// =============================================================================

func TestUiEndpointUrlsPortNormalization(t *testing.T) {
	tests := []struct {
		name      string
		endpoints string
		wantCount int
		validate  func(t *testing.T, urls []*url.URL)
	}{
		{
			name:      "http without port adds :80 variant",
			endpoints: "http://example.com",
			wantCount: 2,
			validate: func(t *testing.T, urls []*url.URL) {
				hosts := []string{urls[0].Host, urls[1].Host}
				assert.Contains(t, hosts, "example.com", "Should contain host without port")
				assert.Contains(t, hosts, "example.com:80", "Should contain host with :80")
			},
		},
		{
			name:      "http with :80 adds variant without port",
			endpoints: "http://example.com:80",
			wantCount: 2,
			validate: func(t *testing.T, urls []*url.URL) {
				hosts := []string{urls[0].Host, urls[1].Host}
				assert.Contains(t, hosts, "example.com", "Should contain host without port")
				assert.Contains(t, hosts, "example.com:80", "Should contain host with :80")
			},
		},
		{
			name:      "https without port adds :443 variant",
			endpoints: "https://example.com",
			wantCount: 2,
			validate: func(t *testing.T, urls []*url.URL) {
				hosts := []string{urls[0].Host, urls[1].Host}
				assert.Contains(t, hosts, "example.com", "Should contain host without port")
				assert.Contains(t, hosts, "example.com:443", "Should contain host with :443")
			},
		},
		{
			name:      "https with :443 adds variant without port",
			endpoints: "https://example.com:443",
			wantCount: 2,
			validate: func(t *testing.T, urls []*url.URL) {
				hosts := []string{urls[0].Host, urls[1].Host}
				assert.Contains(t, hosts, "example.com", "Should contain host without port")
				assert.Contains(t, hosts, "example.com:443", "Should contain host with :443")
			},
		},
		{
			name:      "non-standard ports don't add variants",
			endpoints: "http://example.com:8080",
			wantCount: 1,
			validate: func(t *testing.T, urls []*url.URL) {
				assert.Equal(t, "example.com:8080", urls[0].Host, "Non-standard port should be preserved")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("PHOTOVIEW_SERVE_UI", "0")
			t.Setenv("PHOTOVIEW_UI_ENDPOINTS", tt.endpoints)

			// Call the exported test function to get fresh computation
			urls := utils.ComputeUiEndpointUrlsForTest()

			require.NotNil(t, urls, "ComputeUiEndpointUrlsForTest should return URLs")
			require.Len(t, urls, tt.wantCount, "Endpoint count mismatch")

			if tt.validate != nil {
				tt.validate(t, urls)
			}
		})
	}
}

// =============================================================================
// Integration Tests
// =============================================================================

func TestApiListenUrlUsesApiEndpointPath(t *testing.T) {
	apiEndpoint := mustParseURL("/custom/endpoint")
	listenURL := mustParseURL("http://127.0.0.1:4001/custom/endpoint")

	utils.ConfigureTestEndpoints(apiEndpoint, listenURL, nil)
	t.Cleanup(utils.ResetTestEndpoints)

	listenUrl := utils.ApiListenUrl()
	apiUrl := utils.ApiEndpointUrl()

	assert.Equal(t, apiUrl.Path, listenUrl.Path, "ApiListenUrl should use path from ApiEndpointUrl")
}

func TestApiListenUrlWithAbsoluteApiEndpoint(t *testing.T) {
	apiEndpoint := mustParseURL("https://external.com/api")
	listenURL := mustParseURL("http://127.0.0.1:4001/api")

	utils.ConfigureTestEndpoints(apiEndpoint, listenURL, nil)
	t.Cleanup(utils.ResetTestEndpoints)

	listenUrl := utils.ApiListenUrl()
	assert.Equal(t, "/api", listenUrl.Path, "Should use path from absolute API endpoint")
	assert.Equal(t, "127.0.0.1:4001", listenUrl.Host, "Should use configured listen address")
}

// =============================================================================
// ConfigureTestEndpoints Direct Tests
// =============================================================================

func TestConfigureTestEndpointsDirectly(t *testing.T) {
	apiEndpoint := mustParseURL("/test-api")
	listenURL := mustParseURL("http://10.0.0.1:9000/test-api")
	uiEndpoints := []*url.URL{
		mustParseURL("https://test-ui1.com"),
		mustParseURL("https://test-ui2.com"),
	}

	utils.ConfigureTestEndpoints(apiEndpoint, listenURL, uiEndpoints)
	t.Cleanup(utils.ResetTestEndpoints)

	// Verify endpoints are set correctly
	gotApi := utils.ApiEndpointUrl()
	require.NotNil(t, gotApi)
	assert.Equal(t, "/test-api", gotApi.Path)

	gotListen := utils.ApiListenUrl()
	require.NotNil(t, gotListen)
	assert.Equal(t, "10.0.0.1:9000", gotListen.Host)

	urls := utils.UiEndpointUrls()
	require.Len(t, urls, 2)
	assert.Equal(t, "test-ui1.com", urls[0].Host)
	assert.Equal(t, "test-ui2.com", urls[1].Host)
}

func TestResetTestEndpoints(t *testing.T) {
	// Set test endpoints
	apiEndpoint := mustParseURL("/test")
	utils.ConfigureTestEndpoints(apiEndpoint, nil, nil)

	// Verify they're set
	gotApi := utils.ApiEndpointUrl()
	require.NotNil(t, gotApi)
	assert.Equal(t, "/test", gotApi.Path)

	// Reset
	utils.ResetTestEndpoints()

	// Now set new test endpoints
	apiEndpoint2 := mustParseURL("/production")
	utils.ConfigureTestEndpoints(apiEndpoint2, nil, nil)

	// Should now use the new value
	gotApi2 := utils.ApiEndpointUrl()
	require.NotNil(t, gotApi2)
	assert.Equal(t, "/production", gotApi2.Path)

	// Clean up
	utils.ResetTestEndpoints()
}
