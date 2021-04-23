package models_test

import (
	"testing"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/test_utils"
	"github.com/stretchr/testify/assert"
)

func TestSiteInfo(t *testing.T) {
	db := test_utils.DatabaseTest(t)

	site_info, err := models.GetSiteInfo(db)

	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, models.DefaultSiteInfo(), *site_info)
}
