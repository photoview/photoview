package server

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/photoview/photoview/api/graphql/auth"
	"github.com/photoview/photoview/api/log"
	"github.com/photoview/photoview/api/utils"
	"github.com/wsxiaoys/terminal/color"
)

const maxLogBodyBytes int64 = 50_000

var (
	logFile   *os.File
	logMutex  sync.RWMutex
	logWriter io.Writer
	logLevel  string
)

// InitializeLogging sets up the logging system with optional file output
func InitializeLogging() error {
	logMutex.Lock()
	defer logMutex.Unlock()

	logLevel = strings.ToLower(utils.AccessLogLevel())
	if logLevel == "" {
		logLevel = "info"
	}

	// Default to console output
	logWriter = os.Stdout

	// If log path is configured, open file and create multi-writer
	if logPath := utils.AccessLogPath(); logPath != "" {
		file, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("failed to open access log file %s: %w", logPath, err)
		}
		logFile = file
		logWriter = io.MultiWriter(os.Stdout, file)
		log.Info(context.Background(), "Access logging enabled to file", "logfile", logPath)
	}

	if logLevel == "debug" {
		log.Warn(context.Background(), "Debug access logging enabled")
	}

	return nil
}

// CloseLogging closes logging resources
func CloseLogging() {
	logMutex.Lock()
	defer logMutex.Unlock()

	if logFile != nil {
		if err := logFile.Close(); err != nil {
			log.Error(context.Background(), "Failed to close log file", "error", err)
		}
		logWriter = os.Stdout
		logFile = nil
	}
}

// Thread-safe log writing
func writeLog(format string, args ...interface{}) {
	logMutex.Lock()
	defer logMutex.Unlock()

	if logWriter != nil {
		fmt.Fprintf(logWriter, format, args...)
	}
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		logMutex.RLock()
		debugEnabled := (logLevel == "debug")
		logMutex.RUnlock()

		// Debug logging: incoming request
		if debugEnabled {
			writeLog("\n=== INCOMING REQUEST [%d ms] ===\n", time.Now().UnixMilli()) //TODO: replace with human-readable time
			writeLog("Method: %s\n", r.Method)
			writeLog("URI: %s\n", r.URL.RequestURI())
			writeLog("Host: %s\n", r.Host)
			writeLog("RemoteAddr: %s\n", r.RemoteAddr)
			writeLog("User-Agent: %s\n", r.UserAgent())
			writeLog("Content-Length: %d\n", r.ContentLength)

			// Log all headers
			writeLog("Headers:\n")
			for name, values := range r.Header {
				for _, value := range values {
					if isSensitiveHeader(name) {
						writeLog("  %s: [REDACTED]\n", name)
					} else {
						writeLog("  %s: %s\n", name, value)
					}
				}
			}

			// Log request body (with size limit for safety)
			if r.Body != nil && r.ContentLength > 0 && r.ContentLength < maxLogBodyBytes {
				bodyBytes, err := io.ReadAll(r.Body)
				if err == nil {
					r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
					// Only log text-like content types
					if isTextualContent(r.Header, bodyBytes) {
						writeLog("Body: %s\n", string(bodyBytes))
					} else {
						writeLog("Body: [binary content, %d bytes]\n", len(bodyBytes))
					}
				}
			} else if r.ContentLength >= maxLogBodyBytes {
				writeLog("Body: [large content, %d bytes - not logged]\n", r.ContentLength)
			}
			writeLog("=== PROCESSING ===\n")
		}

		// Choose appropriate response writer based on debug mode
		if debugEnabled {
			// Use debug writer with capture capabilities
			debugWriter := newDebugStatusResponseWriter(&w)
			next.ServeHTTP(debugWriter, r)

			elapsed := time.Since(start)
			elapsedMs := elapsed.Nanoseconds() / 1e6 // Convert to milliseconds

			// Debug logging: response
			writeLog("=== RESPONSE [%d ms] ===\n", time.Now().UnixMilli()) //TODO: replace with human-readable time
			writeLog("Status: %d\n", debugWriter.status)
			writeLog("Duration: %d ms\n", elapsedMs)
			writeLog("Response-Size: %d bytes\n", debugWriter.bodySize)

			// Log response headers if available
			if len(debugWriter.capturedHeaders) > 0 {
				writeLog("Response Headers:\n")
				for name, values := range debugWriter.capturedHeaders {
					for _, value := range values {
						if isSensitiveHeader(name) {
							writeLog("  %s: [REDACTED]\n", name)
						} else {
							writeLog("  %s: %s\n", name, value)
						}
					}
				}
			}

			// Log response body with same binary detection
			if debugWriter.bodyBuffer.Len() > 0 {
				responseBody := debugWriter.bodyBuffer.Bytes()
				if isTextualContent(debugWriter.capturedHeaders, responseBody) {
					writeLog("Response Body: %s\n", string(responseBody))
				} else {
					writeLog("Response Body: [binary content, %d bytes]\n", len(responseBody))
				}
			} else if debugWriter.bodySize > maxLogBodyBytes {
				writeLog("Response Body: [large content, %d bytes - not logged]\n", debugWriter.bodySize)
			} else if debugWriter.bodySize > 0 {
				writeLog("Response Body: [content was written but not captured]\n")
			}
			writeLog("========================\n\n")

			// Standard logging
			logStandardRequest(r, debugWriter.status, elapsedMs)

		} else {
			// Use simple writer with minimal overhead
			simpleWriter := newSimpleStatusResponseWriter(&w)
			next.ServeHTTP(simpleWriter, r)

			elapsed := time.Since(start)
			elapsedMs := elapsed.Nanoseconds() / 1e6

			// Only standard logging (no debug overhead)
			logStandardRequest(r, simpleWriter.status, elapsedMs)
		}
	})
}

func logStandardRequest(r *http.Request, status int, elapsedMs int64) {
	date := time.Now().Format("2006/01/02 15:04:05")
	user := auth.UserFromContext(r.Context())
	requestText := fmt.Sprintf("%s%s", r.Host, r.URL.RequestURI())

	// Color coding for status (preserve existing logic)
	var statusColor string
	switch {
	case status < 200:
		statusColor = color.Colorize("b")
	case status < 300:
		statusColor = color.Colorize("g")
	case status < 400:
		statusColor = color.Colorize("c")
	case status < 500:
		statusColor = color.Colorize("y")
	default:
		statusColor = color.Colorize("r")
	}

	// Color coding for method
	method := r.Method
	var methodColor string
	switch {
	case method == http.MethodGet:
		methodColor = color.Colorize("b")
	case method == http.MethodPost:
		methodColor = color.Colorize("g")
	case method == http.MethodOptions:
		methodColor = color.Colorize("y")
	default:
		methodColor = color.Colorize("r")
	}

	userText := "unauthenticated"
	if user != nil {
		userText = color.Sprintf("@ruser: %s", user.Username)
	}

	statusText := color.Sprintf("%s%s %s%d", methodColor, r.Method, statusColor, status)
	durationText := color.Sprintf("@c%dms", elapsedMs)

	writeLog("%s %s %s %s %s\n", date, statusText, requestText, durationText, userText)
}

func isTextualContent(headers http.Header, body []byte) bool {
	looksTextual := false

	if values, ok := headers["Content-Type"]; ok {
		for _, value := range values {
			value = strings.ToLower(value)
			if strings.HasPrefix(value, "multipart/") ||
				strings.HasPrefix(value, "image/") ||
				strings.HasPrefix(value, "audio/") ||
				strings.HasPrefix(value, "video/") ||
				strings.Contains(value, "octet-stream") {
				return false
			}
			if strings.HasPrefix(value, "text/") ||
				strings.Contains(value, "json") ||
				strings.Contains(value, "xml") ||
				strings.Contains(value, "x-www-form-urlencoded") ||
				strings.Contains(value, "charset=") ||
				strings.Contains(value, "utf-8") {
				looksTextual = true
				break
			}
		}
	}

	if !looksTextual {
		return !isBinaryData(body)
	}
	return looksTextual
}

func isBinaryData(data []byte) bool {
	if len(data) == 0 {
		return false
	}

	// Check first 512 bytes (or less) for null bytes and UTF-8 validity
	checkSize := len(data)
	if checkSize > 512 {
		checkSize = 512
	}
	// Check if it's valid UTF-8 text. If it's not, it's likely binary
	if !utf8.Valid(data[:checkSize]) {
		return true
	}
	// NUL byte is a strong binary signal
	if bytes.IndexByte(data[:checkSize], 0x00) != -1 {
		return true
	}
	return false
}

// Detect sensitive headers
func isSensitiveHeader(name string) bool {
	n := strings.ToLower(name)
	switch n {
	case "authorization", "cookie", "set-cookie",
		"proxy-authorization", "www-authenticate", "proxy-authenticate",
		"x-api-key", "x-auth-token", "x-access-token",
		"x-csrf-token", "x-xsrf-token":
		return true
	}
	// Heuristic patterns
	for _, pat := range []string{
		"api-key", "apikey", "auth", "token", "secret", "password", "session", "bearer", "jwt", "oauth",
	} {
		if strings.Contains(n, pat) {
			return true
		}
	}
	return false
}

type simpleStatusResponseWriter struct {
	http.ResponseWriter
	status   int
	hijacker http.Hijacker
}

// Enhanced status response writer that captures headers
type debugStatusResponseWriter struct {
	http.ResponseWriter
	status          int
	hijacker        http.Hijacker
	capturedHeaders http.Header
	bodyBuffer      *bytes.Buffer
	bodySize        int64
}

func newSimpleStatusResponseWriter(w *http.ResponseWriter) *simpleStatusResponseWriter {
	writer := &simpleStatusResponseWriter{
		ResponseWriter: *w,
	}

	if hj, ok := (*w).(http.Hijacker); ok {
		writer.hijacker = hj
	}

	return writer
}

func newDebugStatusResponseWriter(w *http.ResponseWriter) *debugStatusResponseWriter {
	writer := &debugStatusResponseWriter{
		ResponseWriter:  *w,
		capturedHeaders: make(http.Header),
		bodyBuffer:      &bytes.Buffer{},
	}

	if hj, ok := (*w).(http.Hijacker); ok {
		writer.hijacker = hj
	}

	return writer
}

func (w *simpleStatusResponseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *debugStatusResponseWriter) WriteHeader(status int) {
	w.status = status
	for k, v := range w.ResponseWriter.Header() {
		w.capturedHeaders[k] = v
	}
	w.ResponseWriter.WriteHeader(status)
}

func (w *simpleStatusResponseWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = 200
	}
	return w.ResponseWriter.Write(b)
}

func (w *debugStatusResponseWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = 200
		// Capture headers on implicit WriteHeader
		for k, v := range w.ResponseWriter.Header() {
			w.capturedHeaders[k] = v
		}
	}

	// Capture response body (with size limit for memory safety)
	if w.bodySize < maxLogBodyBytes {
		remainingCapacity := maxLogBodyBytes - w.bodySize
		if int64(len(b)) <= remainingCapacity {
			w.bodyBuffer.Write(b)
		} else {
			// Write only what fits
			w.bodyBuffer.Write(b[:remainingCapacity])
		}
	}
	w.bodySize += int64(len(b))

	return w.ResponseWriter.Write(b)
}

func (w *simpleStatusResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if w.hijacker == nil {
		return nil, nil, errors.New("http.Hijacker not implemented by underlying http.ResponseWriter")
	}
	return w.hijacker.Hijack()
}

func (w *debugStatusResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if w.hijacker == nil {
		return nil, nil, errors.New("http.Hijacker not implemented by underlying http.ResponseWriter")
	}
	return w.hijacker.Hijack()
}
