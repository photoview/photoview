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
}

func (w *authResponseWriter) Write(b []byte) (int, error) {
	if w.authTokenFromResolver != "" {
		http.SetCookie(w, &http.Cookie{
			Name:     "auth-token",
			Value:    w.authTokenFromResolver,
			Path:     "/",
			SameSite: http.SameSiteLaxMode,
			Expires:  time.Now().Add(14 * 24 * time.Hour),
		})
	}
	return w.ResponseWriter.Write(b)
}

func AuthCookieSetter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		arw := authResponseWriter{w, w.(http.Hijacker), ""}
		userIDContextKey := userAccessTokenCtxKey

		ctx := context.WithValue(r.Context(), userIDContextKey, &arw.authTokenFromResolver)
		r = r.WithContext(ctx)

		next.ServeHTTP(&arw, r)
	})
}
