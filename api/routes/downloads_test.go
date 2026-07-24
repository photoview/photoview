package routes

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/test_utils"
	"github.com/stretchr/testify/assert"
)

func TestDownloadRouteRejectsInvalidAlbumIDsBeforeDatabaseLookup(t *testing.T) {
	router := mux.NewRouter()
	RegisterDownloadRoutes(nil, router)

	tests := []struct {
		name    string
		albumID string
	}{
		{name: "text", albumID: "not-a-number"},
		{name: "SQL condition", albumID: "id%3D1%20OR%201%3D1"},
		{name: "time-based SQL condition", albumID: "id%3D1%20AND%20SLEEP%281%29"},
		{name: "zero", albumID: "0"},
		{name: "negative", albumID: "-1"},
		{name: "integer overflow", albumID: "999999999999999999999999999999999999"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/album/"+test.albumID+"/original", nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusBadRequest, rec.Code)
			assert.Equal(t, "invalid album id\n", rec.Body.String())
		})
	}
}

func TestDownloadRouteAcceptsNumericAlbumID(t *testing.T) {
	db := test_utils.DatabaseTest(t)
	album := models.Album{Title: "download-test", Path: t.TempDir()}
	assert.NoError(t, db.Create(&album).Error)

	router := mux.NewRouter()
	RegisterDownloadRoutes(db, router)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/album/%d/original", album.ID), nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
	assert.Contains(t, rec.Body.String(), "unauthorized")
}
