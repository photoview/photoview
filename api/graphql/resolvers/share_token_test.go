package resolvers

import (
	"context"
	"testing"
	"time"

	"github.com/photoview/photoview/api/graphql/models"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	// use memory mode
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}
	err = db.AutoMigrate(&models.ShareToken{})
	if err != nil {
		t.Fatalf("failed to migrate database: %v", err)
	}
	return db
}

func TestShareTokenValidatePassword_SQLite(t *testing.T) {
	password := "123456"
	hashBytes, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	hashedPassword := string(hashBytes)

	now := time.Now()
	expiredTime := now.Add(-24 * time.Hour)
	futureTime := now.Add(24 * time.Hour)

	tests := []struct {
		name        string
		credentials models.ShareTokenCredentials
		prepareData func(db *gorm.DB)
		wantResult  bool
		wantErr     bool
		wantErrMsg  string
	}{
		{
			name: "Case 1: Token not exist",
			credentials: models.ShareTokenCredentials{
				Token: "NOT_EXIST",
			},
			prepareData: func(db *gorm.DB) {
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
			prepareData: func(db *gorm.DB) {
				db.Create(&models.ShareToken{
					Value:  "EXPIRED_TOKEN",
					Expire: &expiredTime,
				})
			},
			wantResult: false,
			wantErr:    true,
			wantErrMsg: "share expired",
		},
		{
			name: "Case 3: correct pass",
			credentials: models.ShareTokenCredentials{
				Token:    "CORRECT_PASS",
				Password: &password,
			},
			prepareData: func(db *gorm.DB) {
				db.Create(&models.ShareToken{
					Value:    "CORRECT_PASS",
					Expire:   &futureTime,
					Password: &hashedPassword,
				})
			},
			wantResult: true,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			if tt.prepareData != nil {
				tt.prepareData(db)
			}
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
