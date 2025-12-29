package models

import (
	"crypto/rand"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type User struct {
	Model
	Username string  `gorm:"unique;size:128"`
	Password *string `gorm:"size:256"`
	// RootPath string  `gorm:"size:512`
	Albums []Album `gorm:"many2many:user_albums;constraint:OnDelete:CASCADE;"`
	Admin  bool    `gorm:"default:false"`
}

type UserMediaData struct {
	ModelTimestamps
	UserID   int  `gorm:"primaryKey;autoIncrement:false"`
	MediaID  int  `gorm:"primaryKey;autoIncrement:false"`
	Favorite bool `gorm:"not null;default:false"`
}

type UserAlbums struct {
	UserID  int `gorm:"primaryKey;autoIncrement:false;constraint:OnDelete:CASCADE;"`
	AlbumID int `gorm:"primaryKey;autoIncrement:false;constraint:OnDelete:CASCADE;"`
}

type AccessToken struct {
	Model
	UserID int       `gorm:"not null;index"`
	User   User      `gorm:"constraint:OnDelete:CASCADE;"`
	Value  string    `gorm:"not null;size:24;index"`
	Expire time.Time `gorm:"not null;index"`
}

type UserPreferences struct {
	Model
	UserID             int  `gorm:"not null;index"`
	User               User `gorm:"constraint:OnDelete:CASCADE;"`
	Language           *LanguageTranslation
	DefaultLandingPage *string `gorm:"size:64"`
}

func (u *UserPreferences) BeforeSave(tx *gorm.DB) error {

	if u.Language != nil && *u.Language == "" {
		u.Language = nil
	}

	if u.Language != nil {
		langStr := string(*u.Language)
		foundMatch := false
		for _, lang := range AllLanguageTranslation {
			if string(lang) == langStr {
				foundMatch = true
				break
			}
		}

		if !foundMatch {
			return errors.New("invalid language value")
		}
	}

	// Validate DefaultLandingPage
	if u.DefaultLandingPage != nil && *u.DefaultLandingPage == "" {
		u.DefaultLandingPage = nil
	}

	if u.DefaultLandingPage != nil {
		validPages := []string{"/timeline", "/albums", "/places", "/people"}
		foundMatch := false
		for _, page := range validPages {
			if *u.DefaultLandingPage == page {
				foundMatch = true
				break
			}
		}

		if !foundMatch {
			return errors.New("invalid default landing page value")
		}
	}

	return nil
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

func RegisterUser(db *gorm.DB, username string, password *string, admin bool) (*User, error) {
	user := User{
		Username: username,
		Admin:    admin,
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

	tokenValue := string(bytes)
	expire := time.Now().Add(14 * 24 * time.Hour)

	token := AccessToken{
		UserID: user.ID,
		Value:  tokenValue,
		Expire: expire,
	}

	result := db.Create(&token)
	if result.Error != nil {
		return nil, errors.Wrap(result.Error, "saving access token to database")
	}

	return &token, nil
}

// FillAlbums fill user.Albums with albums from database
func (user *User) FillAlbums(db *gorm.DB) error {
	// Albums already present
	if len(user.Albums) > 0 {
		return nil
	}

	if err := db.Model(&user).Association("Albums").Find(&user.Albums); err != nil {
		return errors.Wrap(err, "fill user albums")
	}

	return nil
}

func (user *User) OwnsAlbum(db *gorm.DB, album *Album) (bool, error) {
	filter := func(query *gorm.DB) *gorm.DB {
		return query.Where(
			"EXISTS (SELECT 1 FROM user_albums WHERE user_albums.user_id = ? AND user_albums.album_id = id LIMIT 1)",
			user.ID)
	}

	ownedParents, err := album.GetParents(db, filter)
	if err != nil {
		return false, err
	}

	return len(ownedParents) > 0, nil
}

// FavoriteMedia sets/clears a media as favorite for the user
func (user *User) FavoriteMedia(db *gorm.DB, mediaID int, favorite bool) (*Media, error) {
	userMediaData := UserMediaData{
		UserID:   user.ID,
		MediaID:  mediaID,
		Favorite: favorite,
	}

	if err := db.Clauses(clause.OnConflict{UpdateAll: true}).Create(&userMediaData).Error; err != nil {
		return nil, errors.Wrapf(err, "update user favorite media in database")
	}

	var media Media
	if err := db.First(&media, mediaID).Error; err != nil {
		return nil, errors.Wrap(err, "get media from database after favorite update")
	}

	return &media, nil
}
