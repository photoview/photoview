package auth

import (
	context "context"
	http "net/http"
	"time"

	"github.com/felixge/httpsnoop"
)

// AuthCookieSetter returns HTTP middleware that automatically sets an auth-token
// cookie on responses when a GraphQL resolver has produced an access token.
// It stores a pointer to the token field in the request context under
// userAccessTokenCtxKey so resolvers can populate it directly.
// separateDomain should be true when the UI and API are served from different
// origins, which requires SameSite=None;Secure cookie attributes.
func AuthCookieSetter(separateDomain bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var authToken string
			cookieWritten := false

			writeIfNeeded := func() {
				if cookieWritten || authToken == "" {
					return
				}
				sameSite := http.SameSiteLaxMode
				secure := false
				if separateDomain {
					// SameSite=None;Secure is required for cross-origin cookies (e.g. UI and API on different domains)
					sameSite = http.SameSiteNoneMode
					secure = true
				}
				http.SetCookie(w, &http.Cookie{
					Name:     "auth-token",
					Value:    authToken,
					Path:     "/",
					SameSite: sameSite,
					Secure:   secure,
					Expires:  time.Now().Add(14 * 24 * time.Hour),
				})
				cookieWritten = true
			}

			wrapped := httpsnoop.Wrap(w, httpsnoop.Hooks{
				WriteHeader: func(next httpsnoop.WriteHeaderFunc) httpsnoop.WriteHeaderFunc {
					return func(code int) {
						writeIfNeeded()
						next(code)
					}
				},
				Write: func(next httpsnoop.WriteFunc) httpsnoop.WriteFunc {
					return func(b []byte) (int, error) {
						writeIfNeeded()
						return next(b)
					}
				},
			})

			ctx := context.WithValue(r.Context(), userAccessTokenCtxKey, &authToken)
			next.ServeHTTP(wrapped, r.WithContext(ctx))
		})
	}
}
