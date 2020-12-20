package models

import (
	"crypto/rand"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	Model
	Username string  `gorm:"unique,size:128"`
	Password *string `gorm:"size:256`
	// RootPath string  `gorm:"size:512`
	Albums []Album `gorm:"many2many:user_albums"`
	Admin  bool    `gorm:"default:false"`
}

type UserMediaData struct {
	ModelTimestamps
	UserID   int  `gorm:"primaryKey;autoIncrement:false"`
	MediaID  int  `gorm:"primaryKey;autoIncrement:false"`
	Favorite bool `gorm:"not null;default:false"`
}

type AccessToken struct {
	Model
	UserID int       `gorm:"not null"`
	User   User      `gorm:"constraint:OnDelete:CASCADE;"`
	Value  string    `gorm:"not null, size:24`
	Expire time.Time `gorm:"not null"`
}

var ErrorInvalidUserCredentials = errors.New("invalid credentials")

func AuthorizeUser(db *gorm.DB, username string, password string) (*User, error) {
	var user User

	result := db.Where("username = ?", username).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrorInvalidUserCredentials
		}
		return nil, errors.Wrap(result.Error, "failed to get user by username when authorizing")
	}

	if user.Password == nil {
		return nil, errors.New("user does not have a password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*user.Password), []byte(password)); err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return nil, ErrorInvalidUserCredentials
		} else {
			return nil, errors.Wrap(err, "compare user password hash")
		}
	}

	return &user, nil
}

var ErrorInvalidRootPath = errors.New("invalid root path")

func ValidRootPath(rootPath string) bool {
	_, err := os.Stat(rootPath)
	if err != nil {
		log.Printf("Warn: invalid root path: '%s'\n%s\n", rootPath, err)
		return false
	}

	return true
}

func RegisterUser(db *gorm.DB, username string, password *string, admin bool) (*User, error) {
	// if !ValidRootPath(rootPath) {
	// 	return nil, ErrorInvalidRootPath
	// }

	user := User{
		Username: username,
		// RootPath: rootPath,
		Admin: admin,
	}

	if password != nil {
		hashedPassBytes, err := bcrypt.GenerateFromPassword([]byte(*password), 12)
		if err != nil {
			return nil, errors.Wrap(err, "failed to hash password")
		}
		hashedPass := string(hashedPassBytes)

		user.Password = &hashedPass
	}

	result := db.Create(&user)
	if result.Error != nil {
		return nil, errors.Wrap(result.Error, "insert new user with password into database")
	}

	return &user, nil
}

func (user *User) GenerateAccessToken(db *gorm.DB) (*AccessToken, error) {
	bytes := make([]byte, 24)
	if _, err := rand.Read(bytes); err != nil {
		return nil, errors.New(fmt.Sprintf("Could not generate token: %s\n", err.Error()))
	}
	const CHARACTERS = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	for i, b := range bytes {
		bytes[i] = CHARACTERS[b%byte(len(CHARACTERS))]
	}

	token_value := string(bytes)
	expire := time.Now().Add(14 * 24 * time.Hour)
	// expireString := expire.UTC().Format("2006-01-02 15:04:05")

	// if _, err := database.Exec("INSERT INTO access_token (value, expire, user_id) VALUES (?, ?, ?)", token_value, expireString, user.UserID); err != nil {
	// 	return nil, err
	// }

	token := AccessToken{
		UserID: user.ID,
		Value:  token_value,
		Expire: expire,
	}

	result := db.Create(&token)
	if result.Error != nil {
		return nil, errors.Wrap(result.Error, "saving access token to database")
	}

	return &token, nil
}

func VerifyTokenAndGetUser(db *gorm.DB, token string) (*User, error) {

	var accessToken AccessToken
	result := db.Where("expire > ? AND value = ?", time.Now(), token).First(&accessToken)
	if result.Error != nil {
		return nil, result.Error
	}

	var user User
	result = db.First(&user, accessToken.UserID)
	if result.Error != nil {
		return nil, result.Error
	}

	return &user, nil
}
