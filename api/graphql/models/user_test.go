package models_test

import (
	"testing"
	"time"

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

	assert.NotNil(t, user)
	assert.EqualValues(t, "admin", user.Username)
	assert.NotNil(t, user.Password)
	assert.NotEqualValues(t, "1234", user.Password) // should be hashed
	assert.True(t, user.Admin)

	user, err = models.AuthorizeUser(db, "admin", "1234")
	if !assert.NoError(t, err) {
		return
	}

	assert.NotNil(t, user)
	assert.EqualValues(t, "admin", user.Username)

}

func TestAccessToken(t *testing.T) {
	db := test_utils.DatabaseTest(t)

	pass := "<hashed_password>"
	user := models.User{
		Username: "user1",
		Password: &pass,
		Admin:    false,
	}

	if !assert.NoError(t, db.Save(&user).Error) {
		return
	}

	access_token, err := user.GenerateAccessToken(db)
	if !assert.NoError(t, err) {
		return
	}

	assert.NotNil(t, access_token)
	assert.Equal(t, user.ID, access_token.UserID)
	assert.NotEmpty(t, access_token.Value)
	assert.True(t, access_token.Expire.After(time.Now()))
}
