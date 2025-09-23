package server

import (
	"bytes"
	"mime"
	"net/http"
	"strings"
	"unicode/utf8"
)

// Additional utility for enhanced content detection
func detectContentTypeWithFallback(headers http.Header, body []byte) string {
	// Try header first
	contentType := headers.Get("Content-Type")
	if contentType == "" {
		// Fallback to detection from body
		if len(body) > 0 {
			contentType = http.DetectContentType(body)
		}
	}

	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil && mediaType == "" {
		// Fallback to simple string matching for malformed types
		mediaType = strings.ToLower(strings.TrimSpace(contentType))
		if idx := strings.Index(mediaType, ";"); idx != -1 {
			mediaType = strings.TrimSpace(mediaType[:idx])
		}
	}

	return mediaType
}

// isTextualContent determines if content should be logged as text based on MIME type
func isTextualContent(contentType string, body []byte) bool {
	if contentType == "application/octet-stream" {
		return !isBinaryData(body)
	}
	return isTextualMimeType(contentType)
}

// isTextualMimeType checks if a MIME type represents textual content
func isTextualMimeType(contentType string) bool {
	// Explicit binary MIME types (quick rejection)
	binaryTypes := map[string]bool{
		"application/octet-stream":     true,
		"application/zip":              true,
		"application/gzip":             true,
		"application/x-tar":            true,
		"application/x-rar-compressed": true,
		"application/x-7z-compressed":  true,
		"application/x-bzip2":          true,
		"application/pdf":              true,
		"application/vnd.ms-excel":     true,
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":         true,
		"application/vnd.ms-powerpoint":                                             true,
		"application/vnd.openxmlformats-officedocument.presentationml.presentation": true,
		"application/msword": true,
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
	}

	if binaryTypes[contentType] {
		return false
	}

	// Binary media type prefixes
	binaryPrefixes := []string{
		"image/",
		"audio/",
		"video/",
		"font/",
		"application/font-",
	}

	for _, prefix := range binaryPrefixes {
		if strings.HasPrefix(contentType, prefix) {
			return false
		}
	}

	// Special case: multipart with binary content
	if strings.HasPrefix(contentType, "multipart/") &&
		(strings.Contains(contentType, "form-data") || strings.Contains(contentType, "mixed")) {
		return false
	}

	// Textual MIME types (explicit allowlist)
	textualTypes := map[string]bool{
		"text/plain":                        true,
		"text/html":                         true,
		"text/css":                          true,
		"text/javascript":                   true,
		"text/csv":                          true,
		"text/xml":                          true,
		"text/markdown":                     true,
		"text/rtf":                          true,
		"application/json":                  true,
		"application/ld+json":               true,
		"application/xml":                   true,
		"application/soap+xml":              true,
		"application/xhtml+xml":             true,
		"application/rss+xml":               true,
		"application/atom+xml":              true,
		"application/javascript":            true,
		"application/x-javascript":          true,
		"application/x-www-form-urlencoded": true,
		"application/x-yaml":                true,
		"application/yaml":                  true,
		"application/toml":                  true,
		"application/graphql":               true,
		"application/x-httpd-php":           true,
		"application/x-sh":                  true,
	}

	if textualTypes[contentType] {
		return true
	}

	// Text prefix matching (covers text/* family)
	if strings.HasPrefix(contentType, "text/") {
		return true
	}

	// JSON variants (handles vendor-specific JSON types)
	if strings.Contains(contentType, "json") {
		return true
	}

	// XML variants (excludes Office OpenXML formats)
	if strings.Contains(contentType, "xml") && !strings.Contains(contentType, "openxml") {
		return true
	}

	// YAML variants
	if strings.Contains(contentType, "yaml") {
		return true
	}

	return false
}

// Enhanced binary detection with performance optimization
func isBinaryData(data []byte) bool {
	if len(data) == 0 {
		return false
	}

	// Check smaller sample first for performance
	sampleSize := len(data)
	if sampleSize > 512 {
		sampleSize = 512
	}
	sample := data[:sampleSize]

	// Quick null byte check (strong binary indicator)
	if bytes.IndexByte(sample, 0x00) != -1 {
		return true
	}

	// Control character check (excluding common text control chars)
	controlChars := 0
	for _, b := range sample {
		// Allow common text control characters: tab(9), LF(10), CR(13)
		if b < 32 && b != 9 && b != 10 && b != 13 {
			controlChars++
		}
	}

	// If more than 10% are control characters, likely binary
	if controlChars > len(sample)/10 {
		return true
	}

	// UTF-8 validation (most expensive check last)
	return !utf8.Valid(sample)
}
