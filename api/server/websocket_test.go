package server

import (
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/photoview/photoview/api/utils"
	"github.com/stretchr/testify/assert"
)

// configureTestEndpointsFromEnv reads current env vars and configures test endpoints.
// Returns a cleanup function that should be deferred.
func configureTestEndpointsFromEnv(t *testing.T) {
	t.Helper()

	var apiEndpoint *url.URL
	var uiEndpoints []*url.URL

	// Parse API endpoint
	apiEndpointStr := os.Getenv("PHOTOVIEW_API_ENDPOINT")
	if apiEndpointStr == "" {
		apiEndpointStr = "/api"
	}
	var err error
	apiEndpoint, err = url.Parse(apiEndpointStr)
	if err != nil {
		t.Fatalf("Failed to parse API endpoint: %v", err)
	}

	// Parse UI endpoints if not serving UI
	serveUI := os.Getenv("PHOTOVIEW_SERVE_UI")
	if serveUI == "0" || serveUI == "false" {
		uiEndpointsStr := os.Getenv("PHOTOVIEW_UI_ENDPOINTS")
		if uiEndpointsStr != "" {
			for _, urlStr := range strings.Split(uiEndpointsStr, ",") {
				urlStr = strings.TrimSpace(urlStr)
				if urlStr == "" {
					continue
				}
				parsedURL, err := url.Parse(urlStr)
				if err == nil && parsedURL.Scheme != "" && parsedURL.Host != "" {
					uiEndpoints = append(uiEndpoints, parsedURL)

					// Replicate port normalization logic from production (managePort)
					host := parsedURL.Hostname()
					port := parsedURL.Port()
					scheme := parsedURL.Scheme

					switch scheme {
					case "http":
						if port == "" {
							// Add variant with explicit port 80
							withPort, _ := url.Parse("http://" + host + ":80")
							uiEndpoints = append(uiEndpoints, withPort)
						} else if port == "80" {
							// Add variant without port
							withoutPort, _ := url.Parse("http://" + host)
							uiEndpoints = append(uiEndpoints, withoutPort)
						}
					case "https":
						if port == "" {
							// Add variant with explicit port 443
							withPort, _ := url.Parse("https://" + host + ":443")
							uiEndpoints = append(uiEndpoints, withPort)
						} else if port == "443" {
							// Add variant without port
							withoutPort, _ := url.Parse("https://" + host)
							uiEndpoints = append(uiEndpoints, withoutPort)
						}
					}
				}
			}
		}
	}

	utils.ConfigureTestEndpoints(apiEndpoint, nil, uiEndpoints)

	// Cleanup: reset to nil after test
	t.Cleanup(func() {
		utils.ConfigureTestEndpoints(nil, nil, nil)
	})
}

// =============================================================================
// WebsocketUpgrader CheckOrigin Tests
// =============================================================================

func TestWebsocketUpgraderDevModeAllowsAllOrigins(t *testing.T) {
	upgrader := WebsocketUpgrader(true)

	testCases := []struct {
		name   string
		origin string
	}{
		{"no origin header", ""},
		{"localhost origin", "http://localhost:3000"},
		{"external origin", "https://example.com"},
		{"malicious origin", "javascript:alert(1)"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/ws", nil)
			if tc.origin != "" {
				req.Header.Set("Origin", tc.origin)
			}

			result := upgrader.CheckOrigin(req)
			assert.True(t, result, "DevMode should allow all origins")
		})
	}
}

func TestWebsocketUpgraderShouldServeUIAllowsAllOrigins(t *testing.T) {
	t.Setenv("PHOTOVIEW_SERVE_UI", "1")
	configureTestEndpointsFromEnv(t)

	upgrader := WebsocketUpgrader(false)

	testCases := []struct {
		name   string
		origin string
	}{
		{"no origin header", ""},
		{"localhost origin", "http://localhost:3000"},
		{"external origin", "https://example.com"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/ws", nil)
			if tc.origin != "" {
				req.Header.Set("Origin", tc.origin)
			}

			result := upgrader.CheckOrigin(req)
			assert.True(t, result, "When serving UI internally, should allow all origins")
		})
	}
}

func TestWebsocketUpgraderEmptyOriginAllowed(t *testing.T) {
	t.Setenv("PHOTOVIEW_SERVE_UI", "0")
	t.Setenv("PHOTOVIEW_UI_ENDPOINTS", "https://ui.example.com")
	configureTestEndpointsFromEnv(t)

	upgrader := WebsocketUpgrader(false)
	req := httptest.NewRequest("GET", "/ws", nil)
	// Explicitly not setting Origin header

	result := upgrader.CheckOrigin(req)
	assert.True(t, result, "Empty origin header should be allowed")
}

func TestWebsocketUpgraderInvalidOriginRejected(t *testing.T) {
	t.Setenv("PHOTOVIEW_SERVE_UI", "0")
	t.Setenv("PHOTOVIEW_UI_ENDPOINTS", "https://ui.example.com")
	configureTestEndpointsFromEnv(t)

	upgrader := WebsocketUpgrader(false)
	req := httptest.NewRequest("GET", "/ws", nil)
	req.Header.Set("Origin", "://invalid-url")

	result := upgrader.CheckOrigin(req)
	assert.False(t, result, "Invalid origin URL should be rejected")
}

func TestWebsocketUpgraderMatchingSingleEndpointAllowed(t *testing.T) {
	t.Setenv("PHOTOVIEW_SERVE_UI", "0")
	t.Setenv("PHOTOVIEW_UI_ENDPOINTS", "https://ui.example.com")
	configureTestEndpointsFromEnv(t)

	upgrader := WebsocketUpgrader(false)
	req := httptest.NewRequest("GET", "/ws", nil)
	req.Header.Set("Origin", "https://ui.example.com")

	result := upgrader.CheckOrigin(req)
	assert.True(t, result, "Origin matching configured UI endpoint should be allowed")
}

func TestWebsocketUpgraderMatchingMultipleEndpoints(t *testing.T) {
	t.Setenv("PHOTOVIEW_SERVE_UI", "0")
	t.Setenv("PHOTOVIEW_UI_ENDPOINTS", "https://ui1.example.com,https://ui2.example.com,https://ui3.example.com")
	configureTestEndpointsFromEnv(t)

	testCases := []struct {
		name            string
		origin          string
		shouldBeAllowed bool
	}{
		{"first endpoint", "https://ui1.example.com", true},
		{"second endpoint", "https://ui2.example.com", true},
		{"third endpoint", "https://ui3.example.com", true},
		{"first endpoint with path", "https://ui1.example.com/path", true},
		{"non-matching origin", "https://attacker.com", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			upgrader := WebsocketUpgrader(false)
			req := httptest.NewRequest("GET", "/ws", nil)
			req.Header.Set("Origin", tc.origin)

			result := upgrader.CheckOrigin(req)
			assert.Equal(t, tc.shouldBeAllowed, result,
				"Origin %s should be %v", tc.origin, tc.shouldBeAllowed)
		})
	}
}

func TestWebsocketUpgraderNonMatchingOriginRejected(t *testing.T) {
	t.Setenv("PHOTOVIEW_SERVE_UI", "0")
	t.Setenv("PHOTOVIEW_UI_ENDPOINTS", "https://ui.example.com")
	configureTestEndpointsFromEnv(t)

	testCases := []struct {
		name   string
		origin string
	}{
		{"different domain", "https://attacker.com"},
		{"different subdomain", "https://other.example.com"},
		{"different scheme", "http://ui.example.com"},
		{"different port", "https://ui.example.com:8080"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			upgrader := WebsocketUpgrader(false)
			req := httptest.NewRequest("GET", "/ws", nil)
			req.Header.Set("Origin", tc.origin)

			result := upgrader.CheckOrigin(req)
			assert.False(t, result, "Origin %s should be rejected", tc.origin)
		})
	}
}

func TestWebsocketUpgraderPortMatching(t *testing.T) {
	t.Setenv("PHOTOVIEW_SERVE_UI", "0")
	t.Setenv("PHOTOVIEW_UI_ENDPOINTS", "https://ui.example.com:8443")
	configureTestEndpointsFromEnv(t)

	testCases := []struct {
		name            string
		origin          string
		shouldBeAllowed bool
	}{
		{"matching with port", "https://ui.example.com:8443", true},
		{"default https port", "https://ui.example.com", false},
		{"different port", "https://ui.example.com:9000", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			upgrader := WebsocketUpgrader(false)
			req := httptest.NewRequest("GET", "/ws", nil)
			req.Header.Set("Origin", tc.origin)

			result := upgrader.CheckOrigin(req)
			assert.Equal(t, tc.shouldBeAllowed, result,
				"Origin %s should be %v", tc.origin, tc.shouldBeAllowed)
		})
	}
}

func TestWebsocketUpgraderWhitespaceInEndpoints(t *testing.T) {
	t.Setenv("PHOTOVIEW_SERVE_UI", "0")
	t.Setenv("PHOTOVIEW_UI_ENDPOINTS", " https://ui1.example.com , https://ui2.example.com , https://ui3.example.com ")
	configureTestEndpointsFromEnv(t)

	testCases := []string{
		"https://ui1.example.com",
		"https://ui2.example.com",
		"https://ui3.example.com",
	}

	for _, origin := range testCases {
		t.Run(origin, func(t *testing.T) {
			upgrader := WebsocketUpgrader(false)
			req := httptest.NewRequest("GET", "/ws", nil)
			req.Header.Set("Origin", origin)

			result := upgrader.CheckOrigin(req)
			assert.True(t, result, "Should handle whitespace in endpoint list")
		})
	}
}

func TestWebsocketUpgraderCaseInsensitiveHostMatching(t *testing.T) {
	t.Setenv("PHOTOVIEW_SERVE_UI", "0")
	t.Setenv("PHOTOVIEW_UI_ENDPOINTS", "https://UI.EXAMPLE.COM")
	configureTestEndpointsFromEnv(t)

	testCases := []struct {
		name            string
		origin          string
		shouldBeAllowed bool
	}{
		{"uppercase", "https://UI.EXAMPLE.COM", true},
		{"lowercase", "https://ui.example.com", true},
		{"mixed case", "https://Ui.Example.Com", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			upgrader := WebsocketUpgrader(false)
			req := httptest.NewRequest("GET", "/ws", nil)
			req.Header.Set("Origin", tc.origin)

			result := upgrader.CheckOrigin(req)
			assert.Equal(t, tc.shouldBeAllowed, result,
				"Host matching should be case-insensitive per URL spec")
		})
	}
}

func TestWebsocketUpgraderSanitizationOfMaliciousOrigin(t *testing.T) {
	t.Setenv("PHOTOVIEW_SERVE_UI", "0")
	t.Setenv("PHOTOVIEW_UI_ENDPOINTS", "https://ui.example.com")
	configureTestEndpointsFromEnv(t)

	testCases := []struct {
		name   string
		origin string
	}{
		{"newline in host", "https://attacker.com\n.example.com"},
		{"carriage return in host", "https://attacker.com\r.example.com"},
		{"both newline and CR", "https://attacker.com\r\n.example.com"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			upgrader := WebsocketUpgrader(false)
			req := httptest.NewRequest("GET", "/ws", nil)
			req.Header.Set("Origin", tc.origin)

			result := upgrader.CheckOrigin(req)
			assert.False(t, result, "Malicious origin should be rejected")
			// The test passes if it doesn't panic and properly sanitizes the log output
		})
	}
}

func TestWebsocketUpgraderIPv6Endpoints(t *testing.T) {
	t.Setenv("PHOTOVIEW_SERVE_UI", "0")
	t.Setenv("PHOTOVIEW_UI_ENDPOINTS", "https://[::1]:8080,https://[2001:db8::1]:443")
	configureTestEndpointsFromEnv(t)

	testCases := []struct {
		name            string
		origin          string
		shouldBeAllowed bool
	}{
		{"IPv6 localhost with port", "https://[::1]:8080", true},
		{"IPv6 address with port", "https://[2001:db8::1]:443", true},
		{"IPv6 localhost different port", "https://[::1]:9000", false},
		{"IPv4 localhost", "https://127.0.0.1:8080", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			upgrader := WebsocketUpgrader(false)
			req := httptest.NewRequest("GET", "/ws", nil)
			req.Header.Set("Origin", tc.origin)

			result := upgrader.CheckOrigin(req)
			assert.Equal(t, tc.shouldBeAllowed, result,
				"IPv6 origin %s should be %v", tc.origin, tc.shouldBeAllowed)
		})
	}
}

func TestWebsocketUpgraderPathsDoNotAffectMatching(t *testing.T) {
	t.Setenv("PHOTOVIEW_SERVE_UI", "0")
	t.Setenv("PHOTOVIEW_UI_ENDPOINTS", "https://ui.example.com")
	configureTestEndpointsFromEnv(t)

	testCases := []struct {
		name   string
		origin string
	}{
		{"with path", "https://ui.example.com/some/path"},
		{"with query", "https://ui.example.com?query=value"},
		{"with fragment", "https://ui.example.com#fragment"},
		{"with all", "https://ui.example.com/path?query=value#fragment"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			upgrader := WebsocketUpgrader(false)
			req := httptest.NewRequest("GET", "/ws", nil)
			req.Header.Set("Origin", tc.origin)

			result := upgrader.CheckOrigin(req)
			assert.True(t, result, "Path/query/fragment should not affect host matching")
		})
	}
}
