package oidclogin

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestConfigFromEnvironment(t *testing.T) {
	config, err := ConfigFromEnvironment()
	require.NoError(t, err)
	require.Nil(t, config)
	t.Setenv("PHOTOVIEW_OIDC_ISSUER", "https://id.example.test")
	_, err = ConfigFromEnvironment()
	require.Error(t, err)
	t.Setenv("PHOTOVIEW_OIDC_CLIENT_ID", "photoview")
	t.Setenv("PHOTOVIEW_OIDC_REDIRECT_URL", "https://photos.example.test/api/auth/oidc/callback")
	t.Setenv("PHOTOVIEW_OIDC_SUCCESS_URL", "https://photos.example.test/")
	t.Setenv("PHOTOVIEW_OIDC_SUBJECT_MAP", `{"subject-id":"max"}`)
	config, err = ConfigFromEnvironment()
	require.NoError(t, err)
	require.Equal(t, "max", config.SubjectMap["subject-id"])
}

func TestStateIsBrowserBoundAndSingleUse(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.OIDCLoginState{}))
	callbackURL, _ := url.Parse("https://photos.example.test/api/auth/oidc/callback")
	successURL, _ := url.Parse("https://photos.example.test/")
	handler := &Handler{db: db, redirectURL: callbackURL, successURL: successURL, oauth: &oauth2.Config{ClientID: "photoview", RedirectURL: callbackURL.String(), Endpoint: oauth2.Endpoint{AuthURL: "https://id.example.test/authorize"}}}
	mux := http.NewServeMux()
	handler.Register(mux)
	start := httptest.NewRecorder()
	mux.ServeHTTP(start, httptest.NewRequest(http.MethodGet, "/start", nil))
	require.Equal(t, http.StatusSeeOther, start.Code)
	location, err := url.Parse(start.Header().Get("Location"))
	require.NoError(t, err)
	require.Equal(t, "S256", location.Query().Get("code_challenge_method"))
	state := location.Query().Get("state")
	require.NotEmpty(t, state)
	cookie := start.Result().Cookies()[0]
	require.True(t, cookie.Secure)
	require.True(t, cookie.HttpOnly)
	require.Equal(t, state, cookie.Value)
	var record models.OIDCLoginState
	require.NoError(t, db.Where("state_hash = ?", hash(state)).First(&record).Error)

	invalid := httptest.NewRecorder()
	mux.ServeHTTP(invalid, httptest.NewRequest(http.MethodGet, "/callback?state="+state, nil))
	require.Equal(t, http.StatusSeeOther, invalid.Code)
	require.Contains(t, invalid.Header().Get("Location"), "oidc_error=invalid_state")
	firstRequest := httptest.NewRequest(http.MethodGet, "/callback?state="+state, nil)
	firstRequest.AddCookie(cookie)
	first := httptest.NewRecorder()
	mux.ServeHTTP(first, firstRequest)
	require.Equal(t, http.StatusSeeOther, first.Code)
	require.Contains(t, first.Header().Get("Location"), "oidc_error=authorization_failed")
	var count int64
	require.NoError(t, db.Model(&models.OIDCLoginState{}).Where("state_hash = ?", hash(state)).Count(&count).Error)
	require.Zero(t, count)
	secondRequest := httptest.NewRequest(http.MethodGet, "/callback?state="+state, nil)
	secondRequest.AddCookie(cookie)
	second := httptest.NewRecorder()
	mux.ServeHTTP(second, secondRequest)
	require.Equal(t, http.StatusSeeOther, second.Code)
	require.Contains(t, second.Header().Get("Location"), "oidc_error=invalid_state")
}
