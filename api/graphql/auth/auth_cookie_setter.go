package auth

import (
	context "context"
	http "net/http"
	"time"
)

type authResponseWriter struct {
	http.ResponseWriter
	http.Hijacker
	authTokenFromResolver string
	separateDomain        bool
	cookieWritten         bool
}

func (w *authResponseWriter) writeAuthCookieIfNeeded() {
	if w.cookieWritten == false && w.authTokenFromResolver != "" {
		sameSite := http.SameSiteLaxMode
		secure := false
		if w.separateDomain {
			// SameSite=None;Secure is required for cross-origin cookies (e.g. UI and API on different domains)
			sameSite = http.SameSiteNoneMode
			secure = true
		}
		http.SetCookie(w, &http.Cookie{
			Name:     "auth-token",
			Value:    w.authTokenFromResolver,
			Path:     "/",
			SameSite: sameSite,
			Secure:   secure,
			Expires:  time.Now().Add(14 * 24 * time.Hour),
		})
		w.cookieWritten = true
	}
}

func (w *authResponseWriter) WriteHeader(statusCode int) {
	w.writeAuthCookieIfNeeded()
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *authResponseWriter) Write(b []byte) (int, error) {
	w.writeAuthCookieIfNeeded()
	return w.ResponseWriter.Write(b)
}

func AuthCookieSetter(separateDomain bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			arw := authResponseWriter{w, w.(http.Hijacker), "", separateDomain, false}
			userIDContextKey := userAccessTokenCtxKey

			ctx := context.WithValue(r.Context(), userIDContextKey, &arw.authTokenFromResolver)
			r = r.WithContext(ctx)

			next.ServeHTTP(&arw, r)
		})
	}
}
