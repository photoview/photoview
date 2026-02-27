package server

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// CORSMiddleware Tests
// =============================================================================

func TestCORSMiddlewareDevMode(t *testing.T) {
	t.Run("sets origin from request header", func(t *testing.T) {
		middleware := CORSMiddleware(true)
		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, "http://localhost:3000", rec.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "Origin", rec.Header().Get("Vary"))
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("sets CORS headers", func(t *testing.T) {
		middleware := CORSMiddleware(true)
		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("POST", "/test", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, "GET, POST, OPTIONS", rec.Header().Get("Access-Control-Allow-Methods"))
		assert.Equal(t, "authorization, content-type, content-length, TokenPassword", rec.Header().Get("Access-Control-Allow-Headers"))
		assert.Equal(t, "true", rec.Header().Get("Access-Control-Allow-Credentials"))
		assert.Equal(t, "content-length", rec.Header().Get("Access-Control-Expose-Headers"))
	})

	t.Run("handles OPTIONS request", func(t *testing.T) {
		nextCalled := false
		middleware := CORSMiddleware(true)
		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("OPTIONS", "/test", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.False(t, nextCalled, "next handler should not be called for OPTIONS")
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("calls next handler for non-OPTIONS request", func(t *testing.T) {
		nextCalled := false
		middleware := CORSMiddleware(true)
		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("POST", "/test", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.True(t, nextCalled, "next handler should be called for POST")
	})
}

func TestCORSMiddlewareProductionMode(t *testing.T) {
	t.Run("allows request from matching UI endpoint", func(t *testing.T) {
		// Mock utils.UiEndpointUrls to return test endpoints
		testEndpoint, _ := url.Parse("https://example.com")
		originalUiEndpointUrls := uiEndpointUrlsFunc
		uiEndpointUrlsFunc = func() []*url.URL {
			return []*url.URL{testEndpoint}
		}
		defer func() { uiEndpointUrlsFunc = originalUiEndpointUrls }()

		middleware := CORSMiddleware(false)
		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "https://example.com")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, "https://example.com", rec.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "Origin", rec.Header().Get("Vary"))
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("allows request from one of multiple UI endpoints", func(t *testing.T) {
		endpoint1, _ := url.Parse("https://example.com")
		endpoint2, _ := url.Parse("https://app.example.com")
		endpoint3, _ := url.Parse("https://cdn.example.com")

		originalUiEndpointUrls := uiEndpointUrlsFunc
		uiEndpointUrlsFunc = func() []*url.URL {
			return []*url.URL{endpoint1, endpoint2, endpoint3}
		}
		defer func() { uiEndpointUrlsFunc = originalUiEndpointUrls }()

		middleware := CORSMiddleware(false)
		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "https://app.example.com")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, "https://app.example.com", rec.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "Origin", rec.Header().Get("Vary"))
	})

	t.Run("rejects request from non-matching origin", func(t *testing.T) {
		testEndpoint, _ := url.Parse("https://example.com")
		originalUiEndpointUrls := uiEndpointUrlsFunc
		uiEndpointUrlsFunc = func() []*url.URL {
			return []*url.URL{testEndpoint}
		}
		defer func() { uiEndpointUrlsFunc = originalUiEndpointUrls }()

		middleware := CORSMiddleware(false)
		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "https://evil.com")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Empty(t, rec.Header().Get("Access-Control-Allow-Origin"))
		// No CORS methods should be set when origin doesn't match
		assert.Empty(t, rec.Header().Get("Access-Control-Allow-Methods"))
	})

	t.Run("handles nil UI endpoints", func(t *testing.T) {
		originalUiEndpointUrls := uiEndpointUrlsFunc
		uiEndpointUrlsFunc = func() []*url.URL {
			return nil
		}
		defer func() { uiEndpointUrlsFunc = originalUiEndpointUrls }()

		middleware := CORSMiddleware(false)
		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "https://example.com")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Empty(t, rec.Header().Get("Access-Control-Allow-Origin"))
		assert.Empty(t, rec.Header().Get("Access-Control-Allow-Methods"))
	})

	t.Run("handles empty UI endpoints list", func(t *testing.T) {
		originalUiEndpointUrls := uiEndpointUrlsFunc
		uiEndpointUrlsFunc = func() []*url.URL {
			return []*url.URL{}
		}
		defer func() { uiEndpointUrlsFunc = originalUiEndpointUrls }()

		middleware := CORSMiddleware(false)
		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "https://example.com")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Empty(t, rec.Header().Get("Access-Control-Allow-Origin"))
		assert.Empty(t, rec.Header().Get("Access-Control-Allow-Methods"))
	})

	t.Run("handles OPTIONS request in production", func(t *testing.T) {
		testEndpoint, _ := url.Parse("https://example.com")
		originalUiEndpointUrls := uiEndpointUrlsFunc
		uiEndpointUrlsFunc = func() []*url.URL {
			return []*url.URL{testEndpoint}
		}
		defer func() { uiEndpointUrlsFunc = originalUiEndpointUrls }()

		nextCalled := false
		middleware := CORSMiddleware(false)
		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("OPTIONS", "/test", nil)
		req.Header.Set("Origin", "https://example.com")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.False(t, nextCalled, "next handler should not be called for OPTIONS")
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("handles missing origin header", func(t *testing.T) {
		testEndpoint, _ := url.Parse("https://example.com")
		originalUiEndpointUrls := uiEndpointUrlsFunc
		uiEndpointUrlsFunc = func() []*url.URL {
			return []*url.URL{testEndpoint}
		}
		defer func() { uiEndpointUrlsFunc = originalUiEndpointUrls }()

		middleware := CORSMiddleware(false)
		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("GET", "/test", nil)
		// No Origin header set
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		// Should not set CORS origin when no Origin header is present
		assert.Empty(t, rec.Header().Get("Access-Control-Allow-Origin"))
		// CORS methods should not be set when no origin matches
		assert.Empty(t, rec.Header().Get("Access-Control-Allow-Methods"))
		assert.Equal(t, http.StatusOK, rec.Code)
	})
}

// =============================================================================
// setAllowedCORSOrigin Tests
// =============================================================================

func TestSetAllowedCORSOrigin(t *testing.T) {
	testCases := []struct {
		name           string
		uiEndpoints    []*url.URL
		requestOrigin  string
		expectedOrigin string
		expectHeaders  bool
	}{
		{
			name:           "matches single endpoint",
			uiEndpoints:    mustParseURLs("https://example.com"),
			requestOrigin:  "https://example.com",
			expectedOrigin: "https://example.com",
			expectHeaders:  true,
		},
		{
			name:           "matches first of multiple endpoints",
			uiEndpoints:    mustParseURLs("https://example.com", "https://app.example.com"),
			requestOrigin:  "https://example.com",
			expectedOrigin: "https://example.com",
			expectHeaders:  true,
		},
		{
			name:           "matches second of multiple endpoints",
			uiEndpoints:    mustParseURLs("https://example.com", "https://app.example.com"),
			requestOrigin:  "https://app.example.com",
			expectedOrigin: "https://app.example.com",
			expectHeaders:  true,
		},
		{
			name:           "no match returns empty",
			uiEndpoints:    mustParseURLs("https://example.com"),
			requestOrigin:  "https://evil.com",
			expectedOrigin: "",
			expectHeaders:  false,
		},
		{
			name:           "nil endpoints returns empty",
			uiEndpoints:    nil,
			requestOrigin:  "https://example.com",
			expectedOrigin: "",
			expectHeaders:  false,
		},
		{
			name:           "empty endpoints returns empty",
			uiEndpoints:    []*url.URL{},
			requestOrigin:  "https://example.com",
			expectedOrigin: "",
			expectHeaders:  false,
		},
		{
			name:           "empty origin returns empty",
			uiEndpoints:    mustParseURLs("https://example.com"),
			requestOrigin:  "",
			expectedOrigin: "",
			expectHeaders:  false,
		},
		{
			name:           "different port no match",
			uiEndpoints:    mustParseURLs("https://example.com:8080"),
			requestOrigin:  "https://example.com:9090",
			expectedOrigin: "",
			expectHeaders:  false,
		},
		{
			name:           "different scheme no match",
			uiEndpoints:    mustParseURLs("https://example.com"),
			requestOrigin:  "http://example.com",
			expectedOrigin: "",
			expectHeaders:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			if tc.requestOrigin != "" {
				req.Header.Set("Origin", tc.requestOrigin)
			}
			rec := httptest.NewRecorder()

			result := setAllowedCORSOrigin(tc.uiEndpoints, req, rec)

			assert.Equal(t, tc.expectedOrigin, result)
			if tc.expectHeaders {
				assert.Equal(t, tc.expectedOrigin, rec.Header().Get("Access-Control-Allow-Origin"))
				assert.Equal(t, "Origin", rec.Header().Get("Vary"))
			} else {
				assert.Empty(t, rec.Header().Get("Access-Control-Allow-Origin"))
			}
		})
	}
}

// =============================================================================
// findMatchingOrigin Tests
// =============================================================================

func TestFindMatchingOrigin(t *testing.T) {
	testCases := []struct {
		name             string
		requestOrigin    string
		allowedEndpoints []*url.URL
		expectedResult   string
	}{
		{
			name:             "exact match",
			requestOrigin:    "https://example.com",
			allowedEndpoints: mustParseURLs("https://example.com"),
			expectedResult:   "https://example.com",
		},
		{
			name:             "match with port",
			requestOrigin:    "https://example.com:8080",
			allowedEndpoints: mustParseURLs("https://example.com:8080"),
			expectedResult:   "https://example.com:8080",
		},
		{
			name:             "no match different host",
			requestOrigin:    "https://example.com",
			allowedEndpoints: mustParseURLs("https://other.com"),
			expectedResult:   "",
		},
		{
			name:             "no match different scheme",
			requestOrigin:    "http://example.com",
			allowedEndpoints: mustParseURLs("https://example.com"),
			expectedResult:   "",
		},
		{
			name:             "no match different port",
			requestOrigin:    "https://example.com:8080",
			allowedEndpoints: mustParseURLs("https://example.com:9090"),
			expectedResult:   "",
		},
		{
			name:             "match second endpoint",
			requestOrigin:    "https://app.example.com",
			allowedEndpoints: mustParseURLs("https://example.com", "https://app.example.com", "https://cdn.example.com"),
			expectedResult:   "https://app.example.com",
		},
		{
			name:             "invalid URL returns empty",
			requestOrigin:    "://invalid-url",
			allowedEndpoints: mustParseURLs("https://example.com"),
			expectedResult:   "",
		},
		{
			name:             "empty endpoints returns empty",
			requestOrigin:    "https://example.com",
			allowedEndpoints: []*url.URL{},
			expectedResult:   "",
		},
		{
			name:             "path is ignored in matching",
			requestOrigin:    "https://example.com/path/to/resource",
			allowedEndpoints: mustParseURLs("https://example.com"),
			expectedResult:   "https://example.com",
		},
		{
			name:             "query parameters ignored in matching",
			requestOrigin:    "https://example.com?query=param",
			allowedEndpoints: mustParseURLs("https://example.com"),
			expectedResult:   "https://example.com",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := findMatchingOrigin(tc.requestOrigin, tc.allowedEndpoints)
			assert.Equal(t, tc.expectedResult, result)
		})
	}
}

// =============================================================================
// handleCORS Tests
// =============================================================================

func TestHandleCORS(t *testing.T) {
	t.Run("sets CORS headers when enabled", func(t *testing.T) {
		rec := httptest.NewRecorder()

		result := handleCORS(true, rec)

		assert.NotNil(t, result)
		assert.Equal(t, "GET, POST, OPTIONS", rec.Header().Get("Access-Control-Allow-Methods"))
		assert.Equal(t, "authorization, content-type, content-length, TokenPassword", rec.Header().Get("Access-Control-Allow-Headers"))
		assert.Equal(t, "true", rec.Header().Get("Access-Control-Allow-Credentials"))
		assert.Equal(t, "content-length", rec.Header().Get("Access-Control-Expose-Headers"))
	})

	t.Run("does not set CORS headers when disabled", func(t *testing.T) {
		rec := httptest.NewRecorder()

		result := handleCORS(false, rec)

		assert.NotNil(t, result)
		assert.Empty(t, rec.Header().Get("Access-Control-Allow-Methods"))
		assert.Empty(t, rec.Header().Get("Access-Control-Allow-Headers"))
		assert.Empty(t, rec.Header().Get("Access-Control-Allow-Credentials"))
		assert.Empty(t, rec.Header().Get("Access-Control-Expose-Headers"))
	})
}

// =============================================================================
// Helper Functions
// =============================================================================

func mustParseURLs(urls ...string) []*url.URL {
	result := make([]*url.URL, len(urls))
	for i, u := range urls {
		parsed, err := url.Parse(u)
		if err != nil {
			panic(err)
		}
		result[i] = parsed
	}
	return result
}
