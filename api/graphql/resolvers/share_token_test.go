package resolvers

import (
	"context"
	"testing"
	"time"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/test_utils"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestMain(m *testing.M) {
	test_utils.IntegrationTestRun(m)
}

func TestShareTokenValidatePassword(t *testing.T) {
	test_utils.FilesystemTest(t)
	db := test_utils.DatabaseTest(t)
	pass := "1234"
	user, err := models.RegisterUser(db, "test_user", &pass, true)
	if err != nil {
		t.Fatal("register user error:", err)
	}
	hashBytes, _ := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	hashedPassword := string(hashBytes)

	now := time.Now()
	expiredTime := now.Add(-24 * time.Hour)
	expiredTime = time.Date(
		expiredTime.Year(),
		expiredTime.Month(),
		expiredTime.Day(),
		expiredTime.Hour(),
		expiredTime.Minute(),
		expiredTime.Second(),
		0,
		time.UTC,
	)
	futureTime := now.Add(24 * time.Hour)
	futureTime = time.Date(
		futureTime.Year(),
		futureTime.Month(),
		futureTime.Day(),
		futureTime.Hour(),
		futureTime.Minute(),
		futureTime.Second(),
		0,
		time.UTC,
	)

	db.AutoMigrate(&models.ShareToken{})
	testDataList := []models.ShareToken{
		{
			Value:   "EXPIRED_TOKEN",
			OwnerID: user.ID,
			Expire:  &expiredTime,
		},
		{
			Value:    "CORRECT_PASS",
			OwnerID:  user.ID,
			Expire:   &futureTime,
			Password: &hashedPassword,
		},
	}
	db.Create(testDataList)
	tests := []struct {
		name        string
		credentials models.ShareTokenCredentials
		wantResult  bool
		wantErr     bool
		wantErrMsg  string
	}{
		{
			name: "Case 1: Token not exist",
			credentials: models.ShareTokenCredentials{
				Token: "NOT_EXIST",
			},
			wantResult: false,
			wantErr:    true,
			wantErrMsg: "share not found",
		},
		{
			name: "Case 2: Token expired",
			credentials: models.ShareTokenCredentials{
				Token: "EXPIRED_TOKEN",
			},
			wantResult: false,
			wantErr:    true,
			wantErrMsg: "share expired",
		},
		{
			name: "Case 3: correct pass",
			credentials: models.ShareTokenCredentials{
				Token:    "CORRECT_PASS",
				Password: &pass,
			},
			wantResult: true,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &queryResolver{
				Resolver: &Resolver{
					database: db,
				},
			}
			got, err := r.ShareTokenValidatePassword(context.Background(), tt.credentials)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantErrMsg != "" {
					assert.Contains(t, err.Error(), tt.wantErrMsg)
				}
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.wantResult, got)
		})
	}
}
