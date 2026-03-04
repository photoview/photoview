package auth

import (
	context "context"
	http "net/http"
)

type authResponseWriter struct {
	http.ResponseWriter
	http.Hijacker
	userIDToResolver string
	userIDFromCookie string
}

func (w *authResponseWriter) Write(b []byte) (int, error) {
	if w.userIDToResolver != w.userIDFromCookie {
		http.SetCookie(w, &http.Cookie{
			Name:     "auth-token",
			Value:    w.userIDToResolver,
			Path:     "/",
			SameSite: http.SameSiteLaxMode,
		})
	}
	return w.ResponseWriter.Write(b)
}

func AuthCookieSetter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		arw := authResponseWriter{w, w.(http.Hijacker), "", ""}
		userIDContextKey := userAccessTokenCtxKey

		c, _ := r.Cookie("auth-token")
		if c != nil {
			arw.userIDFromCookie = c.Value
			arw.userIDToResolver = c.Value
		}
		ctx := context.WithValue(r.Context(), userIDContextKey, &arw.userIDToResolver)
		r = r.WithContext(ctx)

		next.ServeHTTP(&arw, r)
	})
}
