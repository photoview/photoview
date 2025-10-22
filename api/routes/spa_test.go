package routes

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSpaHandler(t *testing.T) {
	t.Run("valid paths", func(t *testing.T) {
		tempDir := t.TempDir()
		indexPath := filepath.Join(tempDir, "index.html")
		require.NoError(t, os.WriteFile(indexPath, []byte("index content"), 0644))

		handler, err := NewSpaHandler(tempDir, "index.html")
		assert.NoError(t, err)
		assert.NotEmpty(t, handler.staticPath)
		assert.NotEmpty(t, handler.indexPath)
	})

	t.Run("invalid static path", func(t *testing.T) {
		handler, err := NewSpaHandler("/nonexistent/path", "index.html")
		assert.Error(t, err)
		assert.Empty(t, handler.staticPath)
		assert.Empty(t, handler.indexPath)
	})

	t.Run("static path is not a directory", func(t *testing.T) {
		tempDir := t.TempDir()
		filePath := filepath.Join(tempDir, "file.txt")
		require.NoError(t, os.WriteFile(filePath, []byte("content"), 0644))

		handler, err := NewSpaHandler(filePath, "index.html")
		assert.Error(t, err)
		assert.Empty(t, handler.staticPath)
		assert.Empty(t, handler.indexPath)
	})

	t.Run("index path does not exist", func(t *testing.T) {
		tempDir := t.TempDir()
		handler, err := NewSpaHandler(tempDir, "nonexistent.html")
		assert.Error(t, err)
		assert.Empty(t, handler.staticPath)
		assert.Empty(t, handler.indexPath)
	})

	t.Run("index path is a directory", func(t *testing.T) {
		tempDir := t.TempDir()
		indexDir := filepath.Join(tempDir, "indexdir")
		require.NoError(t, os.Mkdir(indexDir, 0755))

		handler, err := NewSpaHandler(tempDir, "indexdir")
		assert.Error(t, err)
		assert.Empty(t, handler.staticPath)
		assert.Empty(t, handler.indexPath)
	})
}

func TestSpaHandler_ServeHTTP(t *testing.T) {
	// Setup test directory structure
	tempDir := t.TempDir()

	// Create index.html
	indexPath := filepath.Join(tempDir, "index.html")
	require.NoError(t, os.WriteFile(indexPath, []byte("<!DOCTYPE html><html>index</html>"), 0644))

	// Create assets directory
	assetsDir := filepath.Join(tempDir, "assets")
	require.NoError(t, os.Mkdir(assetsDir, 0755))

	// Create test files
	jsFile := filepath.Join(assetsDir, "app.js")
	require.NoError(t, os.WriteFile(jsFile, []byte("console.log('app');"), 0644))

	// Create pre-compressed variants
	require.NoError(t, os.WriteFile(jsFile+".br", []byte("compressed-br"), 0644))
	require.NoError(t, os.WriteFile(jsFile+".zst", []byte("compressed-zst"), 0644))
	require.NoError(t, os.WriteFile(jsFile+".gz", []byte("compressed-gz"), 0644))

	// Create an image file (should not be compressed)
	imgFile := filepath.Join(assetsDir, "logo.png")
	require.NoError(t, os.WriteFile(imgFile, []byte("PNG data"), 0644))

	handler, err := NewSpaHandler(tempDir, "index.html")
	require.NoError(t, err)
	require.NotEmpty(t, handler.staticPath)

	t.Run("serve root returns index.html", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "index")
	})

	t.Run("serve existing file", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/assets/app.js", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "console.log")
	})

	t.Run("serve non-existent file returns index.html (SPA routing)", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/some/spa/route", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "index")
	})

	t.Run("path traversal attack prevented", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/../../../etc/passwd", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("in-tree traversal redirected", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/assets/../index.html", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusMovedPermanently, rec.Code)
	})

	t.Run("empty handler configuration returns 500", func(t *testing.T) {
		emptyHandler := SpaHandler{}
		req := httptest.NewRequest("GET", "/", nil)
		rec := httptest.NewRecorder()

		emptyHandler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}

func TestSpaHandler_PrecompressedFiles(t *testing.T) {
	tempDir := t.TempDir()

	// Create index.html
	indexPath := filepath.Join(tempDir, "index.html")
	require.NoError(t, os.WriteFile(indexPath, []byte("index"), 0644))

	// Create assets directory
	assetsDir := filepath.Join(tempDir, "assets")
	require.NoError(t, os.Mkdir(assetsDir, 0755))

	// Create test file with all compression variants
	jsFile := filepath.Join(assetsDir, "app.js")
	jsContent := []byte("console.log('original');")
	require.NoError(t, os.WriteFile(jsFile, jsContent, 0644))
	require.NoError(t, os.WriteFile(jsFile+".br", []byte("br-compressed"), 0644))
	require.NoError(t, os.WriteFile(jsFile+".zst", []byte("zst-compressed"), 0644))
	require.NoError(t, os.WriteFile(jsFile+".gz", []byte("gz-compressed"), 0644))

	handler, err := NewSpaHandler(tempDir, "index.html")
	require.NoError(t, err)

	t.Run("serve brotli when accepted", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/assets/app.js", nil)
		req.Header.Set("Accept-Encoding", "gzip, deflate, br")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "br", rec.Header().Get("Content-Encoding"))
		assert.Contains(t, rec.Header().Get("Vary"), "Accept-Encoding")
		assert.Contains(t, rec.Header().Get("Content-Type"), "javascript")
		assert.Equal(t, "br-compressed", rec.Body.String())
	})

	t.Run("serve zstd when br not accepted", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/assets/app.js", nil)
		req.Header.Set("Accept-Encoding", "gzip, zstd")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "zstd", rec.Header().Get("Content-Encoding"))
		assert.Contains(t, rec.Header().Get("Vary"), "Accept-Encoding")
		assert.Contains(t, rec.Header().Get("Content-Type"), "javascript")
		assert.Equal(t, "zst-compressed", rec.Body.String())
	})

	t.Run("serve gzip when only gzip accepted", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/assets/app.js", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "gzip", rec.Header().Get("Content-Encoding"))
		assert.Contains(t, rec.Header().Get("Vary"), "Accept-Encoding")
		assert.Contains(t, rec.Header().Get("Content-Type"), "javascript")
		assert.Equal(t, "gz-compressed", rec.Body.String())
	})

	t.Run("serve uncompressed when no encoding accepted", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/assets/app.js", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Empty(t, rec.Header().Get("Content-Encoding"))
		assert.Contains(t, rec.Body.String(), "original")
	})

	t.Run("serve original when pre-compressed file missing", func(t *testing.T) {
		// Create file without pre-compressed variants
		cssFile := filepath.Join(assetsDir, "style.css")
		require.NoError(t, os.WriteFile(cssFile, []byte("body { margin: 0; }"), 0644))

		req := httptest.NewRequest("GET", "/assets/style.css", nil)
		req.Header.Set("Accept-Encoding", "gzip, br")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Empty(t, rec.Header().Get("Content-Encoding"))
		assert.Contains(t, rec.Body.String(), "margin")
	})

	t.Run("do not compress already-compressed formats", func(t *testing.T) {
		pngFile := filepath.Join(assetsDir, "logo.png")
		require.NoError(t, os.WriteFile(pngFile, []byte("PNG data"), 0644))
		require.NoError(t, os.WriteFile(pngFile+".br", []byte("should-not-serve"), 0644))

		req := httptest.NewRequest("GET", "/assets/logo.png", nil)
		req.Header.Set("Accept-Encoding", "br")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Empty(t, rec.Header().Get("Content-Encoding"))
		assert.Equal(t, "PNG data", rec.Body.String())
	})
}

func TestSpaHandler_CacheHeaders(t *testing.T) {
	tempDir := t.TempDir()

	indexPath := filepath.Join(tempDir, "index.html")
	require.NoError(t, os.WriteFile(indexPath, []byte("index"), 0644))

	assetsDir := filepath.Join(tempDir, "assets")
	require.NoError(t, os.Mkdir(assetsDir, 0755))

	jsFile := filepath.Join(assetsDir, "app.js")
	require.NoError(t, os.WriteFile(jsFile, []byte("console.log('app');"), 0644))

	otherFile := filepath.Join(tempDir, "manifest.json")
	require.NoError(t, os.WriteFile(otherFile, []byte("{}"), 0644))

	handler, err := NewSpaHandler(tempDir, "index.html")
	require.NoError(t, err)

	t.Run("long-term cache for assets", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/assets/app.js", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "public, max-age=31536000, immutable", rec.Header().Get("Cache-Control"))
	})

	t.Run("short cache with revalidation for non-assets", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/manifest.json", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "public, max-age=3600, must-revalidate", rec.Header().Get("Cache-Control"))
	})

	t.Run("cache headers for index.html", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "public, max-age=3600, must-revalidate", rec.Header().Get("Cache-Control"))
	})
}

func TestSpaHandler_IndexHTML(t *testing.T) {
	tempDir := t.TempDir()

	indexPath := filepath.Join(tempDir, "index.html")
	indexContent := []byte("<!DOCTYPE html><html>index</html>")
	require.NoError(t, os.WriteFile(indexPath, indexContent, 0644))

	// Create pre-compressed index.html
	require.NoError(t, os.WriteFile(indexPath+".br", []byte("index-br-compressed"), 0644))

	handler, err := NewSpaHandler(tempDir, "index.html")
	require.NoError(t, err)

	t.Run("serve pre-compressed index.html when accepted", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Accept-Encoding", "br")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "br", rec.Header().Get("Content-Encoding"))
		assert.Equal(t, "index-br-compressed", rec.Body.String())
	})

	t.Run("serve uncompressed index.html when encoding not accepted", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Empty(t, rec.Header().Get("Content-Encoding"))
		assert.Contains(t, rec.Body.String(), "index")
	})

	t.Run("SPA routing serves index.html for non-existent paths", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/app/users/123", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "index")
	})
}

func TestIsCompressedFormat(t *testing.T) {
	tests := []struct {
		name     string
		ext      string
		expected bool
	}{
		{"png", ".png", true},
		{"jpg", ".jpg", true},
		{"jpeg", ".jpeg", true},
		{"gif", ".gif", true},
		{"webp", ".webp", true},
		{"ico", ".ico", true},
		{"zip", ".zip", true},
		{"gz", ".gz", true},
		{"br", ".br", true},
		{"zst", ".zst", true},
		{"woff", ".woff", true},
		{"woff2", ".woff2", true},
		{"js", ".js", false},
		{"css", ".css", false},
		{"html", ".html", false},
		{"json", ".json", false},
		{"txt", ".txt", false},
		{"uppercase PNG", ".PNG", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isCompressedFormat(tt.ext)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFilterZeroQuality(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no q params (fast path)",
			input:    "gzip, deflate, br",
			expected: "gzip, deflate, br",
		},
		{
			name:     "filter q=0",
			input:    "gzip;q=1.0, deflate;q=0, br;q=0.8",
			expected: "gzip;q=1.0,br;q=0.8",
		},
		{
			name:     "filter q=0.0",
			input:    "gzip;q=0.0, br;q=1",
			expected: "br;q=1",
		},
		{
			name:     "filter with spaces",
			input:    "gzip; q=0.0, deflate; q=1.0, br",
			expected: "deflate; q=1.0,br",
		},
		{
			name:     "all encodings filtered",
			input:    "gzip;q=0, deflate;q=0.0",
			expected: "",
		},
		{
			name:     "empty input",
			input:    "",
			expected: "",
		},
		{
			name:     "q=0 but others present",
			input:    "identity;q=0, gzip, br",
			expected: "gzip,br",
		},
		{
			name:     "multiple spaces",
			input:    "gzip;  q=0,  br;  q=1",
			expected: "br;  q=1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterZeroQuality(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSpaHandler_AcceptEncodingEdgeCases(t *testing.T) {
	tempDir := t.TempDir()

	indexPath := filepath.Join(tempDir, "index.html")
	require.NoError(t, os.WriteFile(indexPath, []byte("index"), 0644))

	assetsDir := filepath.Join(tempDir, "assets")
	require.NoError(t, os.Mkdir(assetsDir, 0755))

	jsFile := filepath.Join(assetsDir, "app.js")
	require.NoError(t, os.WriteFile(jsFile, []byte("original"), 0644))
	require.NoError(t, os.WriteFile(jsFile+".br", []byte("br-compressed"), 0644))
	require.NoError(t, os.WriteFile(jsFile+".gz", []byte("gz-compressed"), 0644))

	handler, err := NewSpaHandler(tempDir, "index.html")
	require.NoError(t, err)

	t.Run("reject encoding with q=0", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/assets/app.js", nil)
		req.Header.Set("Accept-Encoding", "br;q=0, gzip")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		// Should serve gzip since br is rejected with q=0
		assert.Equal(t, "gzip", rec.Header().Get("Content-Encoding"))
	})

	t.Run("case insensitive encoding names", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/assets/app.js", nil)
		req.Header.Set("Accept-Encoding", "BR, GZIP")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "br", rec.Header().Get("Content-Encoding"))
	})

	t.Run("wildcard encoding", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/assets/app.js", nil)
		req.Header.Set("Accept-Encoding", "*")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		// Wildcard doesn't match our specific encodings
		assert.Empty(t, rec.Header().Get("Content-Encoding"))
	})

	t.Run("complex Accept-Encoding with quality values", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/assets/app.js", nil)
		req.Header.Set("Accept-Encoding", "gzip;q=0.8, br;q=1.0, deflate;q=0.5")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		// br has highest quality and should be preferred
		assert.Equal(t, "br", rec.Header().Get("Content-Encoding"))
	})
}

func TestSpaHandler_ContentType(t *testing.T) {
	tempDir := t.TempDir()

	indexPath := filepath.Join(tempDir, "index.html")
	require.NoError(t, os.WriteFile(indexPath, []byte("index"), 0644))

	assetsDir := filepath.Join(tempDir, "assets")
	require.NoError(t, os.Mkdir(assetsDir, 0755))

	// Create files with various extensions
	files := map[string]string{
		"app.js":      "text/javascript",
		"style.css":   "text/css",
		"data.json":   "application/json",
		"image.svg":   "image/svg+xml",
		"unknown.xyz": "", // Unknown extension
	}

	for filename := range files {
		filePath := filepath.Join(assetsDir, filename)
		require.NoError(t, os.WriteFile(filePath, []byte("content"), 0644))
		require.NoError(t, os.WriteFile(filePath+".br", []byte("compressed"), 0644))
	}

	handler, err := NewSpaHandler(tempDir, "index.html")
	require.NoError(t, err)

	for filename, expectedType := range files {
		t.Run(filename, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/assets/"+filename, nil)
			req.Header.Set("Accept-Encoding", "br")
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, "br", rec.Header().Get("Content-Encoding"))

			if expectedType != "" {
				assert.Contains(t, rec.Header().Get("Content-Type"), expectedType)
			}
		})
	}
}
