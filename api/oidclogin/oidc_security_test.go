package oidclogin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync"
	"testing"
	"time"

	oidc "github.com/coreos/go-oidc/v3/oidc"
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestProviderEndpointsRequireHTTPS rejects insecure endpoints in discovery metadata.
func TestProviderEndpointsRequireHTTPS(t *testing.T) {
	for _, insecure := range []string{"authorization_endpoint", "token_endpoint", "jwks_uri"} {
		t.Run(insecure, func(t *testing.T) {
			var provider *httptest.Server
			provider = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				metadata := map[string]string{
					"issuer": provider.URL, "authorization_endpoint": provider.URL + "/authorize",
					"token_endpoint": provider.URL + "/token", "jwks_uri": provider.URL + "/keys",
				}
				metadata[insecure] = "http://id.example.test/insecure"
				w.Header().Set("Content-Type", "application/json")
				require.NoError(t, json.NewEncoder(w).Encode(metadata))
			}))
			defer provider.Close()
			config := &Config{Issuer: provider.URL, ClientID: "photoview", RedirectURL: "https://photos.example.test/api/auth/oidc/callback", SuccessURL: "https://photos.example.test/"}
			_, err := New(oidc.ClientContext(context.Background(), provider.Client()), nil, config, "/api/auth/oidc/callback")
			require.ErrorContains(t, err, "valid HTTPS URLs")
			require.ErrorIs(t, err, ErrDiscovery)
		})
	}
}

// TestMissingStateStorageKeepsLocalLoginAvailable checks the startup fallback.
func TestMissingStateStorageKeepsLocalLoginAvailable(t *testing.T) {
	var provider *httptest.Server
	provider = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		metadata := map[string]string{
			"issuer": provider.URL, "authorization_endpoint": provider.URL + "/authorize",
			"token_endpoint": provider.URL + "/token", "jwks_uri": provider.URL + "/keys",
		}
		w.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode(metadata))
	}))
	defer provider.Close()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	config := &Config{Issuer: provider.URL, ClientID: "photoview", RedirectURL: "https://photos.example.test/api/auth/oidc/callback", SuccessURL: "https://photos.example.test/"}
	_, err = New(oidc.ClientContext(context.Background(), provider.Client()), db, config, "/api/auth/oidc/callback")
	require.ErrorIs(t, err, ErrStateStorage)
}

// TestSuccessURLMustBeUIBase rejects a deep link that would misroute login errors.
func TestSuccessURLMustBeUIBase(t *testing.T) {
	config := &Config{Issuer: "https://id.example.test", ClientID: "photoview", RedirectURL: "https://photos.example.test/api/auth/oidc/callback", SuccessURL: "https://photos.example.test/album/123"}
	_, err := New(context.Background(), nil, config, "/api/auth/oidc/callback")
	require.ErrorContains(t, err, "UI base URL")
}

// TestLoginStartLimits verifies local rejection and the shared database budget.
func TestLoginStartLimits(t *testing.T) {
	h := &Handler{}
	for range localStartsPerSecond {
		require.True(t, h.admitLocal())
	}
	response := httptest.NewRecorder()
	h.start(response, httptest.NewRequest(http.MethodGet, "/start", nil))
	require.Equal(t, http.StatusTooManyRequests, response.Code)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.OIDCLoginState{}, &models.OIDCLoginGate{}))
	require.NoError(t, db.Create(&models.OIDCLoginGate{ID: 1, WindowStart: time.Now(), Starts: globalStartsPerMinute - 1}).Error)
	first := &Handler{db: db}
	second := &Handler{db: db}
	require.NoError(t, first.storeState(&models.OIDCLoginState{StateHash: "first", ExpiresAt: time.Now().Add(time.Minute)}))
	require.ErrorIs(t, second.storeState(&models.OIDCLoginState{StateHash: "second", ExpiresAt: time.Now().Add(time.Minute)}), errLoginLimit)
}

// TestOutstandingStateCap rejects new requests when live states fill the shared cap.
func TestOutstandingStateCap(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.OIDCLoginState{}, &models.OIDCLoginGate{}))
	require.NoError(t, db.Create(&models.OIDCLoginGate{ID: 1, WindowStart: time.Now(), Starts: globalStartsPerMinute - 2}).Error)
	states := make([]models.OIDCLoginState, maxOutstandingStates)
	for i := range states {
		states[i] = models.OIDCLoginState{StateHash: fmt.Sprintf("%064x", i), ExpiresAt: time.Now().Add(time.Minute)}
	}
	require.NoError(t, db.CreateInBatches(states, 100).Error)
	h := &Handler{db: db}
	for range 3 {
		err = h.storeState(&models.OIDCLoginState{StateHash: "next", ExpiresAt: time.Now().Add(time.Minute)})
		require.True(t, errors.Is(err, errLoginLimit))
	}
	var gate models.OIDCLoginGate
	require.NoError(t, db.First(&gate, 1).Error)
	require.Equal(t, uint(globalStartsPerMinute), gate.Starts)
	var count int64
	require.NoError(t, db.Model(&models.OIDCLoginState{}).Count(&count).Error)
	require.Equal(t, int64(maxOutstandingStates), count)
}

// TestConcurrentSQLiteStarts checks that one handler serializes state writes.
func TestConcurrentSQLiteStarts(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "states.db")), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.OIDCLoginState{}, &models.OIDCLoginGate{}))
	require.NoError(t, db.Create(&models.OIDCLoginGate{ID: 1, WindowStart: time.Now()}).Error)
	h := &Handler{db: db}
	var wg sync.WaitGroup
	errs := make(chan error, 2)
	for _, state := range []string{"first", "second"} {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs <- h.storeState(&models.OIDCLoginState{StateHash: state, ExpiresAt: time.Now().Add(time.Minute)})
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		require.NoError(t, err)
	}
}
