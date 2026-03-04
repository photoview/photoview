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

type hijackableRecorder struct {
	*httptest.ResponseRecorder
}

func (h *hijackableRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return nil, nil, nil
}

// Emulate what the AuthorizeUser and InitialSetupWizard resolvers do without depending on them directly
func setResponseAuthCookieHandler(token string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie := auth.ResolverCookieFromContext(r.Context())
		*cookie = token
		w.Write([]byte("ok"))
	})
}

// Emulate an endpoint that is not AuthorizeUser or InitialSetupWizard
func noResponseAuthCookieHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})
}

func TestAuthCookieSetterMiddleware(t *testing.T) {

	testCases := []struct {
		name                  string
		responseAuthCookieVal string
	}{
		{
			name:                  "Login or initial setup endpoint sets auth cookie",
			responseAuthCookieVal: "cookie",
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
				authHandler = auth.AuthCookieSetter(setResponseAuthCookieHandler(tc.responseAuthCookieVal))
			} else {
				authHandler = auth.AuthCookieSetter(noResponseAuthCookieHandler())
			}

			recorder := &hijackableRecorder{httptest.NewRecorder()}
			authHandler.ServeHTTP(recorder, req)

			var authTokenValue string
			for _, cookie := range recorder.Result().Cookies() {
				if cookie.Name == "auth-token" {
					authTokenValue = cookie.Value
					break
				}
			}

			assert.Equal(t, tc.responseAuthCookieVal, authTokenValue)
		})
	}
}
