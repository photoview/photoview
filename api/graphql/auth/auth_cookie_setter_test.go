package auth_test

import (
	"bufio"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/photoview/photoview/api/graphql/auth"
	"github.com/stretchr/testify/assert"
)

// hijackableRecorder wraps httptest.ResponseRecorder with a no-op Hijack method
// so it satisfies the http.Hijacker interface required by authResponseWriter.
type hijackableRecorder struct {
	*httptest.ResponseRecorder
}

// Hijack implements http.Hijacker as a no-op for testing purposes.
func (h *hijackableRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return nil, nil, nil
}

// Emulate what the AuthorizeUser and InitialSetupWizard resolvers do without depending on them directly
func setResponseAuthCookieHandler(token string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie := auth.ResolverCookieFromContext(r.Context())
		*cookie = token
		w.WriteHeader(200)
	})
}

// Emulate an endpoint that is not AuthorizeUser or InitialSetupWizard
func noResponseAuthCookieHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})
}

// TestAuthCookieSetterMiddleware verifies that AuthCookieSetter middleware sets
// the auth-token cookie with the correct value and SameSite/Secure attributes
// when a resolver writes a token, and omits the cookie when no token is set.
func TestAuthCookieSetterMiddleware(t *testing.T) {

	testCases := []struct {
		name                  string
		responseAuthCookieVal string
		uiOnSeparateDomain    bool
	}{
		{
			name:                  "Login or initial setup endpoint sets auth cookie with samesite lax",
			responseAuthCookieVal: "cookie",
			uiOnSeparateDomain:    false,
		},
		{
			name:                  "Login or initial setup endpoint sets auth cookie with samesite none",
			responseAuthCookieVal: "cookie",
			uiOnSeparateDomain:    true,
		},
		{
			name:                  "Non Login or initial setup endpoint does not set auth cookie",
			responseAuthCookieVal: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/graphql", nil)

			var authHandler http.Handler

			if tc.responseAuthCookieVal != "" {
				authHandler = auth.AuthCookieSetter(tc.uiOnSeparateDomain)(setResponseAuthCookieHandler(tc.responseAuthCookieVal))
			} else {
				authHandler = auth.AuthCookieSetter(tc.uiOnSeparateDomain)(noResponseAuthCookieHandler())
			}

			recorder := &hijackableRecorder{httptest.NewRecorder()}
			authHandler.ServeHTTP(recorder, req)

			var authToken *http.Cookie
			for _, cookie := range recorder.Result().Cookies() {
				if cookie.Name == "auth-token" {
					authToken = cookie
					break
				}
			}

			if tc.responseAuthCookieVal != "" {
				assert.Equal(t, tc.responseAuthCookieVal, authToken.Value)
				if tc.uiOnSeparateDomain {
					assert.Equal(t, http.SameSiteNoneMode, authToken.SameSite)
					assert.True(t, authToken.Secure)
				} else {
					assert.Equal(t, http.SameSiteLaxMode, authToken.SameSite)
					assert.False(t, authToken.Secure)
				}
			} else {
				assert.Nil(t, authToken)
			}
		})
	}
}
