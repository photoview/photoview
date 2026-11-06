// Package oidclogin adds optional OpenID Connect login without changing local
// password authentication or Photoview's existing user permissions.
package oidclogin

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	oidc "github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/photoview/photoview/api/graphql/models"
)

const stateCookie = "photoview-oidc-state"
const stateLifetime = 5 * time.Minute
const localStartsPerSecond = 20
const globalStartsPerMinute = 300
const maxOutstandingStates = 1600

var errLoginLimit = errors.New("OIDC login start limit reached")

// ErrDiscovery identifies a temporary provider failure; local login should
// remain available even when the identity provider is offline at startup.
var ErrDiscovery = errors.New("OIDC provider discovery unavailable")

// ErrStateStorage identifies unavailable OIDC state storage at startup.
var ErrStateStorage = errors.New("OIDC state storage unavailable")

type Config struct {
	Issuer       string
	ClientID     string
	ClientSecret string
	RedirectURL  string
	SuccessURL   string
	// SubjectMap maps provider subject IDs to existing Photoview usernames.
	SubjectMap map[string]string
}

type Handler struct {
	db          *gorm.DB
	oauth       *oauth2.Config
	verifier    *oidc.IDTokenVerifier
	redirectURL *url.URL
	successURL  *url.URL
	subjects    map[string]string
	limitMu     sync.Mutex
	limitSince  time.Time
	limitStarts int
	storeMu     sync.Mutex
}

// ConfigFromEnvironment returns nil when OIDC is not configured. A partially
// configured provider is an error so an operator does not get a broken button.
func ConfigFromEnvironment() (*Config, error) {
	issuer := os.Getenv("PHOTOVIEW_OIDC_ISSUER")
	clientID := os.Getenv("PHOTOVIEW_OIDC_CLIENT_ID")
	clientSecret := os.Getenv("PHOTOVIEW_OIDC_CLIENT_SECRET")
	redirect := os.Getenv("PHOTOVIEW_OIDC_REDIRECT_URL")
	success := os.Getenv("PHOTOVIEW_OIDC_SUCCESS_URL")
	mapping := os.Getenv("PHOTOVIEW_OIDC_SUBJECT_MAP")
	if issuer == "" && clientID == "" && clientSecret == "" && redirect == "" && success == "" && mapping == "" {
		return nil, nil
	}
	if issuer == "" || clientID == "" || redirect == "" || success == "" || mapping == "" {
		return nil, errors.New("OIDC requires issuer, client ID, redirect URL, success URL, and subject map")
	}
	subjects := map[string]string{}
	if err := json.Unmarshal([]byte(mapping), &subjects); err != nil {
		return nil, fmt.Errorf("invalid OIDC subject map: %w", err)
	}
	if len(subjects) == 0 {
		return nil, errors.New("OIDC subject map cannot be empty")
	}
	linkedUsers := make(map[string]bool, len(subjects))
	for sub, username := range subjects {
		if strings.TrimSpace(sub) == "" || strings.TrimSpace(username) == "" {
			return nil, errors.New("OIDC subject map cannot contain empty subjects or usernames")
		}
		if linkedUsers[username] {
			return nil, errors.New("OIDC subject map cannot link multiple subjects to one user")
		}
		linkedUsers[username] = true
	}
	return &Config{issuer, clientID, clientSecret, redirect, success, subjects}, nil
}

// New validates the OIDC URLs, discovers the provider, and prepares shared
// state admission before enabling SSO.
// A nil config leaves local password login as the only sign-in method.
func New(ctx context.Context, db *gorm.DB, config *Config, callbackPath string) (*Handler, error) {
	if config == nil {
		return nil, nil
	}
	issuerURL, err := url.Parse(config.Issuer)
	if err != nil || issuerURL.Scheme != "https" || issuerURL.Host == "" || issuerURL.User != nil || issuerURL.RawQuery != "" || issuerURL.Fragment != "" {
		return nil, errors.New("OIDC issuer must be an HTTPS URL")
	}
	redirectURL, err := url.Parse(config.RedirectURL)
	if err != nil || redirectURL.Scheme != "https" || redirectURL.Host == "" || redirectURL.User != nil || redirectURL.RawQuery != "" || redirectURL.Fragment != "" || redirectURL.Path != callbackPath {
		return nil, fmt.Errorf("OIDC redirect URL must be HTTPS and end at %s", callbackPath)
	}
	successURL, err := url.Parse(config.SuccessURL)
	if err != nil || successURL.Scheme != "https" || successURL.Host != redirectURL.Host || successURL.User != nil || successURL.RawQuery != "" || successURL.Fragment != "" || (successURL.Path != "" && !strings.HasSuffix(successURL.Path, "/")) {
		return nil, errors.New("OIDC success URL must be an HTTPS UI base URL ending in / on the same host as the callback")
	}
	provider, err := oidc.NewProvider(ctx, config.Issuer)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDiscovery, err)
	}
	var metadata struct {
		JWKSURI string `json:"jwks_uri"`
	}
	if err := provider.Claims(&metadata); err != nil {
		return nil, fmt.Errorf("%w: invalid provider metadata: %v", ErrDiscovery, err)
	}
	endpoint := provider.Endpoint()
	if !secureEndpoint(endpoint.AuthURL) || !secureEndpoint(endpoint.TokenURL) || !secureEndpoint(metadata.JWKSURI) {
		return nil, fmt.Errorf("%w: provider authorization, token, and JWKS endpoints must be valid HTTPS URLs", ErrDiscovery)
	}
	if !db.Migrator().HasTable(&models.OIDCLoginState{}) || !db.Migrator().HasTable(&models.OIDCLoginGate{}) {
		return nil, ErrStateStorage
	}
	if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&models.OIDCLoginGate{ID: 1, WindowStart: time.Now()}).Error; err != nil {
		return nil, fmt.Errorf("%w: %v", ErrStateStorage, err)
	}
	return &Handler{
		db:          db,
		oauth:       &oauth2.Config{ClientID: config.ClientID, ClientSecret: config.ClientSecret, RedirectURL: config.RedirectURL, Endpoint: endpoint, Scopes: []string{oidc.ScopeOpenID}},
		verifier:    provider.Verifier(&oidc.Config{ClientID: config.ClientID}),
		redirectURL: redirectURL,
		successURL:  successURL,
		subjects:    config.SubjectMap,
	}, nil
}

// secureEndpoint accepts an HTTPS provider endpoint without embedded credentials.
func secureEndpoint(raw string) bool {
	u, err := url.Parse(raw)
	return err == nil && u.Scheme == "https" && u.Host != "" && u.User == nil && u.Fragment == "" && u.Opaque == ""
}

// randomString returns 32 cryptographically random bytes as a URL-safe value.
// It is used for browser state and the ID-token nonce.
func randomString() (string, error) {
	value := make([]byte, 32)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(value), nil
}
func hash(value string) string {
	digest := sha256.Sum256([]byte(value))
	return hex.EncodeToString(digest[:])
}
func (h *Handler) callbackPath() string { return h.redirectURL.Path }

func (h *Handler) loginError(w http.ResponseWriter, r *http.Request, code string) {
	loginURL := *h.successURL
	loginURL.Path = strings.TrimRight(loginURL.Path, "/") + "/login"
	loginURL.RawQuery = "oidc_error=" + url.QueryEscape(code)
	http.Redirect(w, r, loginURL.String(), http.StatusSeeOther)
}

func (h *Handler) Register(router *http.ServeMux) {
	router.HandleFunc("GET /status", h.status)
	router.HandleFunc("GET /start", h.start)
	router.HandleFunc("GET /callback", h.callback)
}

func (h *Handler) status(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"enabled":true}`))
}

// start persists a short-lived authorization request and binds its state to
// the browser before redirecting to the provider with PKCE and a nonce.
func (h *Handler) start(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	if !h.admitLocal() {
		http.Error(w, "OIDC login requests are temporarily limited", http.StatusTooManyRequests)
		return
	}
	state, err := randomString()
	if err != nil {
		log.Printf("OIDC login start: random state unavailable: %v", err)
		h.loginError(w, r, "unavailable")
		return
	}
	nonce, err := randomString()
	if err != nil {
		log.Printf("OIDC login start: random nonce unavailable: %v", err)
		h.loginError(w, r, "unavailable")
		return
	}
	verifier := oauth2.GenerateVerifier()
	record := models.OIDCLoginState{StateHash: hash(state), CodeVerifier: verifier, Nonce: nonce, ExpiresAt: time.Now().Add(stateLifetime)}
	if err := h.storeState(&record); err != nil {
		if errors.Is(err, errLoginLimit) {
			http.Error(w, "OIDC login requests are temporarily limited", http.StatusTooManyRequests)
			return
		}
		log.Printf("OIDC login start: state storage failed: %v", err)
		h.loginError(w, r, "unavailable")
		return
	}
	http.SetCookie(w, &http.Cookie{Name: stateCookie, Value: state, Path: h.callbackPath(), MaxAge: int(stateLifetime.Seconds()), Secure: true, HttpOnly: true, SameSite: http.SameSiteLaxMode})
	authURL := h.oauth.AuthCodeURL(state, oauth2.S256ChallengeOption(verifier), oauth2.SetAuthURLParam("nonce", nonce))
	http.Redirect(w, r, authURL, http.StatusSeeOther)
}

// admitLocal bounds requests before any database work on this replica.
func (h *Handler) admitLocal() bool {
	h.limitMu.Lock()
	defer h.limitMu.Unlock()
	now := time.Now()
	if now.Sub(h.limitSince) >= time.Second {
		h.limitSince = now
		h.limitStarts = 0
	}
	if h.limitStarts >= localStartsPerSecond {
		return false
	}
	h.limitStarts++
	return true
}

// storeState serializes a shared rate limit and outstanding-state cap before
// inserting an authorization request, so concurrent replicas use one budget.
func (h *Handler) storeState(record *models.OIDCLoginState) error {
	// SQLite ignores SELECT FOR UPDATE, so serialize this process's requests too.
	h.storeMu.Lock()
	defer h.storeMu.Unlock()
	limited := false
	err := h.db.Transaction(func(tx *gorm.DB) error {
		var gate models.OIDCLoginGate
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&gate, 1).Error; err != nil {
			return err
		}
		now := time.Now()
		if now.Sub(gate.WindowStart) >= time.Minute {
			gate.WindowStart = now
			gate.Starts = 0
		}
		if gate.Starts >= globalStartsPerMinute {
			return errLoginLimit
		}
		if err := tx.Model(&gate).Updates(map[string]interface{}{"window_start": gate.WindowStart, "starts": gate.Starts + 1}).Error; err != nil {
			return err
		}
		if err := tx.Where("expires_at < ?", now).Delete(&models.OIDCLoginState{}).Error; err != nil {
			return err
		}
		var outstanding int64
		if err := tx.Model(&models.OIDCLoginState{}).Count(&outstanding).Error; err != nil {
			return err
		}
		if outstanding >= maxOutstandingStates {
			// Commit the gate charge even though no state is inserted; otherwise
			// repeated cap rejections would bypass the shared rate budget.
			limited = true
			return nil
		}
		return tx.Create(record).Error
	})
	if err != nil {
		return err
	}
	if limited {
		return errLoginLimit
	}
	return nil
}

// callback consumes the browser-bound state once, verifies the ID token, and
// issues a new Photoview session for an existing, explicitly mapped user.
func (h *Handler) callback(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	cookie, err := r.Cookie(stateCookie)
	http.SetCookie(w, &http.Cookie{Name: stateCookie, Path: h.callbackPath(), MaxAge: -1, Secure: true, HttpOnly: true, SameSite: http.SameSiteLaxMode})
	state := r.URL.Query().Get("state")
	if err != nil || len(state) != 43 || subtle.ConstantTimeCompare([]byte(cookie.Value), []byte(state)) != 1 {
		log.Printf("OIDC login callback rejected: invalid browser state")
		h.loginError(w, r, "invalid_state")
		return
	}
	var record models.OIDCLoginState
	if err := h.db.Where("state_hash = ? AND expires_at > ?", hash(state), time.Now()).First(&record).Error; err != nil {
		log.Printf("OIDC login callback rejected: missing or expired state: %v", err)
		h.loginError(w, r, "invalid_state")
		return
	}
	// Consume state before exchange; a callback can only be used once.
	if result := h.db.Where("id = ? AND state_hash = ?", record.ID, record.StateHash).Delete(&models.OIDCLoginState{}); result.Error != nil || result.RowsAffected != 1 {
		log.Printf("OIDC login callback rejected: state already consumed or database error: %v", result.Error)
		h.loginError(w, r, "invalid_state")
		return
	}
	if providerError := r.URL.Query().Get("error"); providerError != "" {
		log.Printf("OIDC provider rejected authorization: %q", providerError)
		h.loginError(w, r, "authorization_failed")
		return
	}
	code := r.URL.Query().Get("code")
	if code == "" {
		log.Printf("OIDC login callback rejected: authorization code missing")
		h.loginError(w, r, "authorization_failed")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()
	token, err := h.oauth.Exchange(ctx, code, oauth2.VerifierOption(record.CodeVerifier))
	if err != nil {
		log.Printf("OIDC authorization code exchange failed: %v", err)
		h.loginError(w, r, "authorization_failed")
		return
	}
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		log.Printf("OIDC authorization code exchange returned no ID token")
		h.loginError(w, r, "authorization_failed")
		return
	}
	idToken, err := h.verifier.Verify(ctx, rawIDToken)
	if err != nil || subtle.ConstantTimeCompare([]byte(idTokenNonce(idToken)), []byte(record.Nonce)) != 1 {
		log.Printf("OIDC ID token verification or nonce check failed: %v", err)
		h.loginError(w, r, "authorization_failed")
		return
	}
	username, ok := h.subjects[idToken.Subject]
	if !ok {
		log.Printf("OIDC login rejected: unlinked subject %q", idToken.Subject)
		h.loginError(w, r, "not_linked")
		return
	}
	var user models.User
	if err := h.db.Where("username = ?", username).First(&user).Error; err != nil {
		log.Printf("OIDC login rejected: mapped Photoview user %q unavailable: %v", username, err)
		h.loginError(w, r, "not_linked")
		return
	}
	accessToken, err := user.GenerateAccessToken(h.db)
	if err != nil {
		log.Printf("OIDC login failed to create Photoview session: %v", err)
		h.loginError(w, r, "unavailable")
		return
	}
	// Photoview's existing UI reads auth-token for GraphQL WebSocket auth. The
	// cookie follows the same session model as local password login.
	http.SetCookie(w, &http.Cookie{Name: "auth-token", Value: accessToken.Value, Path: "/", Expires: accessToken.Expire, Secure: true, SameSite: http.SameSiteLaxMode})
	http.Redirect(w, r, h.successURL.String(), http.StatusSeeOther)
}

// idTokenNonce reads the nonce claim from an ID token, returning an empty
// value when the token or its claims cannot be read so the nonce check fails.
func idTokenNonce(token *oidc.IDToken) string {
	if token == nil {
		return ""
	}
	var claims struct {
		Nonce string `json:"nonce"`
	}
	if err := token.Claims(&claims); err != nil {
		return ""
	}
	return claims.Nonce
}
