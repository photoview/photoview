package auth_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/photoview/photoview/api/dataloader"
	"github.com/photoview/photoview/api/graphql/auth"
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/test_utils"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	test_utils.UnitTestRun(m)
}

func TestTokenFromBearer(t *testing.T) {

	testsValues := []struct {
		name   string
		bearer string
		out    string
		valid  bool
	}{
		{"Valid bearer", "Bearer ZY9YfxFa3TapSAD37XUBFryo", "ZY9YfxFa3TapSAD37XUBFryo", true},
		{"Case insensitive bearer", "bEaReR ZY9YfxFa3TapSAD37XUBFryo", "ZY9YfxFa3TapSAD37XUBFryo", true},
		{"Missing bearer start", "ZY9YfxFa3TapSAD37XUBFryo", "", false},
		{"Empty input", "", "", false},
		{"Invalid token value", "Bearer THIS_IS_INVALID", "", false},
	}

	for _, test := range testsValues {
		t.Run(test.name, func(t *testing.T) {
			token, err := auth.TokenFromBearer(&test.bearer)
			if test.valid {
				assert.NoError(t, err)
				assert.NotNil(t, token)
				assert.Equal(t, test.out, *token)
			} else {
				assert.Error(t, err)
				assert.Nil(t, token)
			}
		})
	}
}

func TestAuthWebsocketInit(t *testing.T) {
	db := test_utils.DatabaseTest(t)

	// Create test user and token
	password := "test_password"
	user, err := models.RegisterUser(db, "testuser", &password, false)
	assert.NoError(t, err)

	token, err := user.GenerateAccessToken(db)
	assert.NoError(t, err)

	testCases := []struct {
		name         string
		initPayload  transport.InitPayload
		expectError  bool
		expectUser   bool
		expectNilCtx bool
	}{
		{
			name:        "Valid authorization",
			initPayload: transport.InitPayload{"Authorization": "Bearer " + token.Value},
			expectError: false,
			expectUser:  true,
		},
		{
			name:        "Missing authorization",
			initPayload: transport.InitPayload{},
			expectError: false,
			expectUser:  false,
		},
		{
			name:         "Invalid bearer format",
			initPayload:  transport.InitPayload{"Authorization": "InvalidFormat"},
			expectError:  true,
			expectNilCtx: true,
		},
		{
			name:         "Invalid token",
			initPayload:  transport.InitPayload{"Authorization": "Bearer INVALID_TOKEN_123456"},
			expectError:  true,
			expectNilCtx: true,
		},
		{
			name:         "Empty token",
			initPayload:  transport.InitPayload{"Authorization": "Bearer "},
			expectError:  true,
			expectNilCtx: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			initFunc := auth.AuthWebsocketInit(db)
			ctx := context.Background()
			req := httptest.NewRequest("GET", "/", nil)
			req = req.WithContext(ctx)

			var contextWithLoaders context.Context
			handler := dataloader.Middleware(db)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				contextWithLoaders = r.Context()
			}))
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, req)

			resultCtx, ackPayload, err := initFunc(contextWithLoaders, tc.initPayload)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tc.expectNilCtx {
				assert.Nil(t, resultCtx)
			} else {
				assert.NotNil(t, resultCtx)
			}

			// Verify InitPayload acknowledgment is always nil (as per PR implementation)
			assert.Nil(t, ackPayload)

			if tc.expectUser {
				retrievedUser := auth.UserFromContext(resultCtx)
				assert.NotNil(t, retrievedUser)
				assert.Equal(t, user.ID, retrievedUser.ID)
				assert.Equal(t, "testuser", retrievedUser.Username)
			} else if !tc.expectNilCtx {
				retrievedUser := auth.UserFromContext(resultCtx)
				assert.Nil(t, retrievedUser)
			}
		})
	}
}

func TestMiddleware(t *testing.T) {
	db := test_utils.DatabaseTest(t)

	// Create test user and token
	password := "test_password"
	user, err := models.RegisterUser(db, "testuser", &password, false)
	assert.NoError(t, err)

	token, err := user.GenerateAccessToken(db)
	assert.NoError(t, err)

	testCases := []struct {
		name         string
		cookieValue  string
		setCookie    bool
		expectStatus int
		expectUser   bool
	}{
		{
			name:         "Valid token cookie",
			cookieValue:  token.Value,
			setCookie:    true,
			expectStatus: 200,
			expectUser:   true,
		},
		{
			name:         "No cookie",
			setCookie:    false,
			expectStatus: 200,
			expectUser:   false,
		},
		{
			name:         "Invalid token",
			cookieValue:  "INVALID_TOKEN",
			setCookie:    true,
			expectStatus: 401,
			expectUser:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/graphql", nil)
			if tc.setCookie {
				req.AddCookie(&http.Cookie{
					Name:  "auth-token",
					Value: tc.cookieValue,
				})
			}

			var capturedContext context.Context
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedContext = r.Context()
			})

			authHandler := auth.Middleware(db)(handler)
			fullHandler := dataloader.Middleware(db)(authHandler)

			recorder := httptest.NewRecorder()
			fullHandler.ServeHTTP(recorder, req)

			assert.Equal(t, tc.expectStatus, recorder.Code)

			if tc.expectUser {
				retrievedUser := auth.UserFromContext(capturedContext)
				assert.NotNil(t, retrievedUser)
				assert.Equal(t, user.ID, retrievedUser.ID)
			} else if recorder.Code == 200 {
				// Handler was called, verify no user in context
				retrievedUser := auth.UserFromContext(capturedContext)
				assert.Nil(t, retrievedUser)
			}
		})
	}
}

func TestContextUserOperations(t *testing.T) {
	db := test_utils.DatabaseTest(t)

	password := "test"
	user, err := models.RegisterUser(db, "testuser", &password, false)
	assert.NoError(t, err)

	t.Run("AddUserToContext and UserFromContext", func(t *testing.T) {
		ctx := context.Background()

		// Initially no user
		retrieved := auth.UserFromContext(ctx)
		assert.Nil(t, retrieved)

		// Add user
		ctxWithUser := auth.AddUserToContext(ctx, user)

		// Retrieve user
		retrieved = auth.UserFromContext(ctxWithUser)
		assert.NotNil(t, retrieved)
		assert.Equal(t, user.ID, retrieved.ID)
		assert.Equal(t, "testuser", retrieved.Username)
	})
}
