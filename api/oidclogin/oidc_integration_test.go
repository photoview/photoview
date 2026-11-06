package oidclogin

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	oidc "github.com/coreos/go-oidc/v3/oidc"
	jose "github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/photoview/photoview/api/graphql/models"
)

// TestSignedOIDCCallbackIssuesExistingUserSession verifies that a signed ID
// token maps to an existing user and creates a Photoview session.
func TestSignedOIDCCallbackIssuesExistingUserSession(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.User{}, &models.AccessToken{}, &models.OIDCLoginState{}))
	user := models.User{Username: "max", Admin: true}
	require.NoError(t, db.Create(&user).Error)
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.RS256, Key: key}, (&jose.SignerOptions{}).WithHeader("kid", "test"))
	require.NoError(t, err)
	jwks, err := json.Marshal(jose.JSONWebKeySet{Keys: []jose.JSONWebKey{{Key: &key.PublicKey, KeyID: "test", Algorithm: "RS256", Use: "sig"}}})
	require.NoError(t, err)
	var provider *httptest.Server
	var nonce string
	provider = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/.well-known/openid-configuration":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{"issuer": provider.URL, "authorization_endpoint": provider.URL + "/authorize", "token_endpoint": provider.URL + "/token", "jwks_uri": provider.URL + "/keys"})
		case "/keys":
			w.Header().Set("Content-Type", "application/json")
			w.Write(jwks)
		case "/token":
			require.NoError(t, r.ParseForm())
			require.NotEmpty(t, r.Form.Get("code_verifier"))
			claims := map[string]interface{}{"iss": provider.URL, "sub": "keycloak-max-id", "aud": "photoview", "exp": time.Now().Add(time.Hour).Unix(), "iat": time.Now().Unix(), "nonce": nonce}
			idToken, signErr := jwt.Signed(signer).Claims(claims).Serialize()
			require.NoError(t, signErr)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{"access_token": "provider-access-token", "token_type": "Bearer", "expires_in": 3600, "id_token": idToken})
		default:
			http.NotFound(w, r)
		}
	}))
	defer provider.Close()
	cfg := &Config{Issuer: provider.URL, ClientID: "photoview", RedirectURL: "https://photos.example.test/api/auth/oidc/callback", SuccessURL: "https://photos.example.test/", SubjectMap: map[string]string{"keycloak-max-id": "max"}}
	handler, err := New(oidc.ClientContext(context.Background(), provider.Client()), db, cfg, "/api/auth/oidc/callback")
	require.NoError(t, err)
	mux := http.NewServeMux()
	handler.Register(mux)
	start := httptest.NewRecorder()
	mux.ServeHTTP(start, httptest.NewRequest(http.MethodGet, "/start", nil))
	require.Equal(t, http.StatusSeeOther, start.Code)
	authURL, err := url.Parse(start.Header().Get("Location"))
	require.NoError(t, err)
	nonce = authURL.Query().Get("nonce")
	require.NotEmpty(t, nonce)
	state := authURL.Query().Get("state")
	callback := httptest.NewRequest(http.MethodGet, "/callback?state="+url.QueryEscape(state)+"&code=good", nil)
	callback = callback.WithContext(oidc.ClientContext(callback.Context(), provider.Client()))
	callback.AddCookie(start.Result().Cookies()[0])
	result := httptest.NewRecorder()
	mux.ServeHTTP(result, callback)
	require.Equal(t, http.StatusSeeOther, result.Code)
	require.Equal(t, cfg.SuccessURL, result.Header().Get("Location"))
	var sessionCookie *http.Cookie
	for _, cookie := range result.Result().Cookies() {
		if cookie.Name == "auth-token" {
			sessionCookie = cookie
		}
	}
	require.NotNil(t, sessionCookie)
	require.True(t, sessionCookie.Secure)
	require.Equal(t, http.SameSiteLaxMode, sessionCookie.SameSite)
	require.Equal(t, "/", sessionCookie.Path)
	var session models.AccessToken
	require.NoError(t, db.Where("value = ?", sessionCookie.Value).First(&session).Error)
	require.Equal(t, user.ID, session.UserID)
	require.Equal(t, int64(1), countSessions(t, db))
}
func countSessions(t *testing.T, db *gorm.DB) int64 {
	var count int64
	require.NoError(t, db.Model(&models.AccessToken{}).Count(&count).Error)
	return count
}
