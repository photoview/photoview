package auth

import (
	context "context"
	http "net/http"
	"time"
)

// authResponseWriter wraps http.ResponseWriter to intercept response writes and
// inject an auth-token cookie when a token has been set by a GraphQL resolver.
// It also embeds http.Hijacker to support WebSocket upgrades.
type authResponseWriter struct {
	http.ResponseWriter
	http.Hijacker
	authTokenFromResolver string
	separateDomain        bool
	cookieWritten         bool
}

// writeAuthCookieIfNeeded sets the auth-token cookie on the response if a token
// was provided by a resolver and the cookie has not already been written.
// For cross-origin deployments (separateDomain=true), it uses SameSite=None;Secure.
func (w *authResponseWriter) writeAuthCookieIfNeeded() {
	if w.cookieWritten || w.authTokenFromResolver == "" {
		return
	}

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

// WriteHeader intercepts the status code write to ensure the auth cookie is set
// before headers are flushed to the client.
func (w *authResponseWriter) WriteHeader(statusCode int) {
	w.writeAuthCookieIfNeeded()
	w.ResponseWriter.WriteHeader(statusCode)
}

// Write intercepts the response body write to ensure the auth cookie is set
// before any body bytes are sent, in case WriteHeader was not called explicitly.
func (w *authResponseWriter) Write(b []byte) (int, error) {
	w.writeAuthCookieIfNeeded()
	return w.ResponseWriter.Write(b)
}

// AuthCookieSetter returns HTTP middleware that automatically sets an auth-token
// cookie on responses when a GraphQL resolver has produced an access token.
// It stores a pointer to the token field in the request context under
// userAccessTokenCtxKey so resolvers can populate it directly.
// separateDomain should be true when the UI and API are served from different
// origins, which requires SameSite=None;Secure cookie attributes.
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
