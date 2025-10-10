package server

import (
	"bufio"
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

type contextKey string

var (
	logFile          io.WriteCloser
	logWriter        io.Writer
	logMutex         sync.RWMutex
	logGlobalContext context.Context
)

// InitializeLogging sets up the logging system with optional file output
func InitializeLogging() {
	logMutex.Lock()
	defer logMutex.Unlock()

	logGlobalContext = context.Background()

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

		logGlobalContext = context.WithValue(logGlobalContext, contextKey("logFile"), logFile)
		logGlobalContext = context.WithValue(logGlobalContext, contextKey("logWriter"), logWriter)

		log.Info(
			logGlobalContext,
			"Access logging enabled to file",
			"logfile", rotatingLogger.Filename,
			"max size in MB", rotatingLogger.MaxSize,
			"max files", rotatingLogger.MaxBackups,
			"max age in days", rotatingLogger.MaxAge,
			"compressed", rotatingLogger.Compress,
		)
	}
}

// CloseLogging closes logging resources
func CloseLogging() {
	logMutex.Lock()
	defer logMutex.Unlock()

	if logFile != nil {
		if err := logFile.Close(); err != nil {
			log.Error(logGlobalContext, "Failed to close log file", "error", err)
		}
		logWriter = os.Stdout
		logFile = nil
		logGlobalContext = context.Background()
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

		// Use simple writer with minimal overhead
		simpleWriter := newStatusResponseWriter(&w)
		next.ServeHTTP(simpleWriter, r)

		elapsed := time.Since(start)
		elapsedMs := elapsed.Nanoseconds() / 1e6

		// Only standard logging (no debug overhead)
		logStandardRequest(r, simpleWriter.status, elapsedMs)
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

type statusResponseWriter struct {
	http.ResponseWriter
	status   int
	hijacker http.Hijacker
}

func newStatusResponseWriter(w *http.ResponseWriter) *statusResponseWriter {
	writer := &statusResponseWriter{
		ResponseWriter: *w,
	}

	if hj, ok := (*w).(http.Hijacker); ok {
		writer.hijacker = hj
	}

	return writer
}

func (w *statusResponseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *statusResponseWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = 200
	}
	return w.ResponseWriter.Write(b)
}

func (w *statusResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if w.hijacker == nil {
		return nil, nil, errors.New("http.Hijacker not implemented by underlying http.ResponseWriter")
	}
	return w.hijacker.Hijack()
}

func (w *statusResponseWriter) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (w *statusResponseWriter) Push(target string, opts *http.PushOptions) error {
	if p, ok := w.ResponseWriter.(http.Pusher); ok {
		return p.Push(target, opts)
	}
	return http.ErrNotSupported
}
