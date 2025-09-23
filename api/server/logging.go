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
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/photoview/photoview/api/graphql/auth"
	"github.com/photoview/photoview/api/log"
	"github.com/photoview/photoview/api/utils"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	logFile         io.WriteCloser
	logMutex        sync.RWMutex
	logWriter       io.Writer
	logLevel        string
	maxLogBodyBytes int64
)

// InitializeLogging sets up the logging system with optional file output
func InitializeLogging() error {
	logMutex.Lock()
	defer logMutex.Unlock()

	logLevel = strings.ToLower(utils.AccessLogLevel())
	if logLevel == "" {
		logLevel = "info"
	}

	maxLogBodyBytes = utils.AccessLogMaxBodyBytes()

	// Default to console output
	logWriter = os.Stdout

	// If log path is configured, set up rotating file logger as part of multi-writer
	if logPath := utils.AccessLogPath(); logPath != "" {
		rotatingLogger := &lumberjack.Logger{
			Filename:   logPath,
			MaxSize:    utils.AccessLogMaxSize(),
			MaxBackups: utils.AccessLogMaxFiles(),
			MaxAge:     utils.AccessLogMaxDays(),
			Compress:   utils.EnvAccessLogIsCompressed.GetBool(),
			LocalTime:  true,
		}

		logFile = rotatingLogger
		logWriter = io.MultiWriter(os.Stdout, logFile)
		log.Info(
			context.Background(),
			"Access logging enabled to file",
			"logfile", logPath,
			"max size in MB", utils.AccessLogMaxSize(),
			"max files", utils.AccessLogMaxFiles(),
			"max age in days", utils.AccessLogMaxDays(),
			"compressed", utils.EnvAccessLogIsCompressed.GetBool(),
		)
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
			var logBuf bytes.Buffer

			logBuf.WriteString(fmt.Sprintf(
				"\n===v INCOMING REQUEST [%s] v===\n",
				time.Now().Format("2006 Jan 02, 15:04:05.000 (MST) -07:00"),
			))
			logBuf.WriteString(fmt.Sprintf("Method: %s\n", r.Method))
			logBuf.WriteString(fmt.Sprintf("URI: %s\n", sanitizeURI(r.URL)))
			logBuf.WriteString(fmt.Sprintf("Host: %s\n", r.Host))
			logBuf.WriteString(fmt.Sprintf("RemoteAddr: %s\n", r.RemoteAddr))
			logBuf.WriteString(fmt.Sprintf("User-Agent: %s\n", r.UserAgent()))
			logBuf.WriteString(fmt.Sprintf("Content-Length: %d\n", r.ContentLength))

			// Log all headers
			logBuf.WriteString("Headers:\n")
			for name, values := range r.Header {
				for _, value := range values {
					if isSensitiveHeader(name) {
						logBuf.WriteString(fmt.Sprintf("  %s: [REDACTED]\n", name))
					} else {
						logBuf.WriteString(fmt.Sprintf("  %s: %s\n", name, value))
					}
				}
			}

			// Log request body (with size limit for safety)
			if r.Body != nil && r.ContentLength > 0 {
				var bodyBytes []byte
				var readErr error

				bodyBytes, readErr = io.ReadAll(io.LimitReader(r.Body, maxLogBodyBytes))

				if readErr != nil {
					logBuf.WriteString(fmt.Sprintf("Body: [error reading request body: %v]\n", readErr))
					if closeErr := r.Body.Close(); closeErr != nil {
						log.Error(context.Background(), "Failed to close request body after read error", "error", closeErr)
					}
					r.Body = io.NopCloser(strings.NewReader(""))
				} else {
					if len(bodyBytes) == 0 {
						logBuf.WriteString("Body: [empty]\n")
					} else {
						// Log based on content analysis
						contentType := detectContentTypeWithFallback(r.Header, bodyBytes)
						if isTextualContent(contentType, bodyBytes) {
							logBuf.WriteString(fmt.Sprintf("Body: %s\n", string(bodyBytes)))
						} else {
							logBuf.WriteString(
								fmt.Sprintf("Body: [binary content, %d bytes, type: %s]\n", len(bodyBytes), contentType),
							)
						}

						// Indicate if content was truncated
						if r.ContentLength > maxLogBodyBytes {
							logBuf.WriteString(fmt.Sprintf("Body: [Note: logged first %d of %d total bytes]\n",
								len(bodyBytes), r.ContentLength))
						}
					}

					// Restore body for downstream handlers
					if r.ContentLength <= maxLogBodyBytes {
						r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
					} else {
						// Large body: restore read portion + remaining unread portion
						// Calculate remaining bytes to read
						remainingBytes := r.ContentLength - int64(len(bodyBytes))
						remaining := io.LimitReader(r.Body, remainingBytes)
						r.Body = io.NopCloser(io.MultiReader(bytes.NewReader(bodyBytes), remaining))
					}
				}
			} else if r.Body != nil && r.ContentLength == 0 {
				logBuf.WriteString("Body: WARN [is not empty, but Content-Length: 0]\n")
			}

			logBuf.WriteString("===^ INCOMING REQUEST ^===\n")

			// Use debug writer with capture capabilities
			debugWriter := newDebugStatusResponseWriter(&w)
			next.ServeHTTP(debugWriter, r)

			elapsed := time.Since(start)
			elapsedMs := elapsed.Nanoseconds() / 1e6 // Convert to milliseconds

			// Debug logging: response
			logBuf.WriteString(fmt.Sprintf(
				"===v RESPONSE [%s] v===\n",
				time.Now().Format("2006 Jan 02, 15:04:05.000 (MST) -07:00"),
			))
			logBuf.WriteString(fmt.Sprintf("Status: %d\n", debugWriter.status))
			logBuf.WriteString(fmt.Sprintf("Duration: %d ms\n", elapsedMs))
			logBuf.WriteString(fmt.Sprintf("Response-Size: %d bytes\n", debugWriter.bodySize))

			// Log response headers if available
			if len(debugWriter.capturedHeaders) > 0 {
				logBuf.WriteString("Response Headers:\n")
				for name, values := range debugWriter.capturedHeaders {
					for _, value := range values {
						if isSensitiveHeader(name) {
							logBuf.WriteString(fmt.Sprintf("  %s: [REDACTED]\n", name))
						} else {
							logBuf.WriteString(fmt.Sprintf("  %s: %s\n", name, value))
						}
					}
				}
			}

			// Log response body with same binary detection
			if debugWriter.bodyBuffer.Len() > 0 {
				responseBody := debugWriter.bodyBuffer.Bytes()
				contentType := detectContentTypeWithFallback(debugWriter.capturedHeaders, responseBody)
				if isTextualContent(contentType, responseBody) {
					logBuf.WriteString(fmt.Sprintf("Response Body: %s\n", string(responseBody)))
				} else {
					logBuf.WriteString(fmt.Sprintf("Response Body: [binary content, %d bytes, type: %s]\n",
						len(responseBody), contentType))
				}
			} else if debugWriter.bodySize > maxLogBodyBytes {
				logBuf.WriteString(fmt.Sprintf(
					"Response Body: [large content, %d bytes - not logged]\n",
					debugWriter.bodySize,
				))
			} else if debugWriter.bodySize > 0 {
				logBuf.WriteString("Response Body: [content was written but not captured]\n")
			}
			logBuf.WriteString("===^ RESPONSE ^===\n\n")
			writeLog("%s", logBuf.String())
			logBuf.Reset()

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
	date := time.Now().Format("2006 Jan 02, 15:04:05 (MST) -07:00")
	user := auth.UserFromContext(r.Context())
	requestText := fmt.Sprintf("%s%s", r.Host, sanitizeURI(r.URL))

	userText := "unauthenticated"
	if user != nil {
		userText = fmt.Sprintf("@ruser: %s", user.Username)
	}

	statusText := fmt.Sprintf("%s %d", r.Method, status)
	durationText := fmt.Sprintf("@c%dms", elapsedMs)

	writeLog("%s %s %s %s %s\n", date, statusText, requestText, durationText, userText)
}

// sanitizeURL redacts sensitive query parameters in a URL
func sanitizeURI(url *url.URL) string {
	if url == nil {
		return ""
	}
	cloneUrl := *url
	queryString := cloneUrl.Query()
	if len(queryString) == 0 {
		return cloneUrl.RequestURI()
	}

	sensitiveKeys := []string{
		"access_token", "token", "auth", "authorization", "apikey", "api_key",
		"password", "passwd", "secret", "signature", "session", "jwt", "code",
	}

	for name := range queryString {
		lowerName := strings.ToLower(name)
		for _, sensitive := range sensitiveKeys {
			if lowerName == sensitive || strings.Contains(lowerName, sensitive) {
				queryString[name] = []string{"[REDACTED]"}
				break
			}
		}
	}
	cloneUrl.RawQuery = queryString.Encode()
	return cloneUrl.RequestURI()
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
		w.capturedHeaders[k] = append([]string(nil), v...)
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
			w.capturedHeaders[k] = append([]string(nil), v...)
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

func (w *simpleStatusResponseWriter) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (w *debugStatusResponseWriter) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (w *simpleStatusResponseWriter) Push(target string, opts *http.PushOptions) error {
	if p, ok := w.ResponseWriter.(http.Pusher); ok {
		return p.Push(target, opts)
	}
	return http.ErrNotSupported
}

func (w *debugStatusResponseWriter) Push(target string, opts *http.PushOptions) error {
	if p, ok := w.ResponseWriter.(http.Pusher); ok {
		return p.Push(target, opts)
	}
	return http.ErrNotSupported
}
