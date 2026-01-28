package server

import (
	"bufio"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoview/photoview/api/graphql/auth"
	"github.com/photoview/photoview/api/graphql/models"
)

func TestLoggingMiddleware(t *testing.T) {

	tests := []struct {
		name          string
		withHealthHdr bool
		withUser      bool
	}{
		{
			name:          "Health check skips logging",
			withHealthHdr: true,
			withUser:      false,
		},
		{
			name:          "Unauthenticated request",
			withHealthHdr: false,
			withUser:      false,
		},
		{
			name:          "Authenticated request",
			withHealthHdr: false,
			withUser:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			req := httptest.NewRequest(http.MethodGet, "/graphql", nil)

			if tt.withHealthHdr {
				req.Header.Set("X-Health-Check", "true")
			}

			if tt.withUser {
				user := &models.User{
					Username: "testuser",
				}
				ctx := auth.AddUserToContext(req.Context(), user)
				req = req.WithContext(ctx)
			}

			handlerCalled := false
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handlerCalled = true
				w.WriteHeader(http.StatusOK)
			})

			recorder := &hijackableRecorder{
				ResponseRecorder: httptest.NewRecorder(),
			}
			middleware := LoggingMiddleware(handler)

			middleware.ServeHTTP(recorder, req)

			assert.True(t, handlerCalled, "handler should be called")
			assert.Equal(t, http.StatusOK, recorder.Code)
		})
	}
}

type hijackableRecorder struct {
	*httptest.ResponseRecorder
}

func (h *hijackableRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return nil, nil, errors.New("hijack not supported in test")
}
