package routes

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gorilla/mux"
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/test_utils"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func requestAlbumDownload(db *gorm.DB, albumID string) *httptest.ResponseRecorder {
	router := mux.NewRouter()
	RegisterDownloadRoutes(db, router)

	req := httptest.NewRequest(http.MethodGet, "/album/"+albumID+"/original", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	return rec
}

func TestDownloadRouteRejectsInvalidAlbumIDsBeforeDatabaseLookup(t *testing.T) {
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
			rec := requestAlbumDownload(nil, test.albumID)

			assert.Equal(t, http.StatusBadRequest, rec.Code)
			assert.Equal(t, "invalid album id\n", rec.Body.String())
		})
	}
}

func TestDownloadRouteAcceptsNumericAlbumID(t *testing.T) {
	db := test_utils.DatabaseTest(t)
	album := models.Album{Title: "download-test", Path: t.TempDir()}
	assert.NoError(t, db.Create(&album).Error)

	rec := requestAlbumDownload(db, strconv.Itoa(album.ID))

	assert.Equal(t, http.StatusForbidden, rec.Code)
	assert.Contains(t, rec.Body.String(), "unauthorized")
}

func TestDownloadRouteReturnsNotFoundForMissingAlbum(t *testing.T) {
	db := test_utils.DatabaseTest(t)

	rec := requestAlbumDownload(db, "1")

	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Equal(t, "404", rec.Body.String())
}

func TestDownloadRouteReturnsInternalServerErrorForDatabaseFailure(t *testing.T) {
	db := test_utils.DatabaseTest(t)
	ctx := context.Background()

	// Force First to return a real query error without closing the shared test pool.
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get database pool: %v", err)
	}
	connection, err := sqlDB.Conn(ctx)
	if err != nil {
		t.Fatalf("get dedicated database connection: %v", err)
	}
	if err := connection.Close(); err != nil {
		t.Fatalf("close dedicated database connection: %v", err)
	}

	failingDB := db.WithContext(ctx)
	failingDB.Statement.ConnPool = connection

	rec := requestAlbumDownload(failingDB, "1")

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Equal(t, internalServerError, rec.Body.String())
}
