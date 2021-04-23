package models_test

import (
	"testing"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/test_utils"
	"github.com/stretchr/testify/assert"
)

func TestUserRegistrationAuthorization(t *testing.T) {
	db := test_utils.DatabaseTest(t)

	password := "1234"
	user, err := models.RegisterUser(db, "admin", &password, true)

	if !assert.NoError(t, err) {
		return
	}

	assert.EqualValues(t, "admin", user.Username)
	assert.NotNil(t, user.Password)
	assert.NotEqualValues(t, "1234", user.Password) // should be hashed
	assert.True(t, user.Admin)
}
