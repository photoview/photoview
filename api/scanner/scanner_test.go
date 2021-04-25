package scanner_test

import (
	"os"
	"testing"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner"
	"github.com/photoview/photoview/api/test_utils"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	os.Exit(test_utils.IntegrationTestRun(m))
}

func TestFullScan(t *testing.T) {
	test_utils.FilesystemTest(t)
	db := test_utils.DatabaseTest(t)

	pass := "1234"
	user, err := models.RegisterUser(db, "test_user", &pass, true)
	if !assert.NoError(t, err) {
		return
	}

	root_album := models.Album{
		Title: "root album",
		Path:  "./test_data",
	}

	if !assert.NoError(t, db.Save(&root_album).Error) {
		return
	}

	err = db.Model(user).Association("Albums").Append(&root_album)
	if !assert.NoError(t, err) {
		return
	}

	if !assert.NoError(t, scanner.InitializeScannerQueue(db)) {
		return
	}

	if !assert.NoError(t, scanner.AddUserToQueue(user)) {
		return
	}

	// wait for all jobs to finish
	scanner.CloseScannerQueue()

	var all_media []*models.Media
	if !assert.NoError(t, db.Find(&all_media).Error) {
		return
	}

	assert.Equal(t, 10, len(all_media))

}
