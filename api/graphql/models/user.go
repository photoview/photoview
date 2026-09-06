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

// UserAlbumData stores per-user, per-album personal display preferences that
// must never affect another user's access to the album - currently just
// whether the viewer has hidden it from their own navigation. Distinct from
// UserAlbums, which is an actual access grant.
type UserAlbumData struct {
	ModelTimestamps
	UserID  int  `gorm:"primaryKey;autoIncrement:false"`
	AlbumID int  `gorm:"primaryKey;autoIncrement:false"`
	Hidden  bool `gorm:"not null;default:false"`
}

type UserAlbums struct {
	UserID  int                  `gorm:"primaryKey;autoIncrement:false;constraint:OnDelete:CASCADE;"`
	AlbumID int                  `gorm:"primaryKey;autoIncrement:false;constraint:OnDelete:CASCADE;"`
	// Default matches AlbumPermissionLevelRead's wire value: a fail-safe
	// default for any row ever inserted without an explicit level.
	Level AlbumPermissionLevel `gorm:"not null;default:READ"`
	// GrantedByUserID is nil for an admin-configured volume/folder grant (an
	// "owner"), or the granting owner's user ID for a peer share. Only a nil
	// GrantedByUserID entitles a user to share the album further (see
	// actions.ShareAlbum) — a share recipient can use their access but
	// cannot re-share it to a fourth person.
	GrantedByUserID *int `gorm:"index"`
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
	UserID   int  `gorm:"not null;index"`
	User     User `gorm:"constraint:OnDelete:CASCADE;"`
	Language *LanguageTranslation
	// SearchResultLimit is the maximum number of albums/media returned per category by a search.
	// nil means the server default is used, 0 means no limit (return all results).
	SearchResultLimit *int
	// ShowAlbumTree controls whether the album tree sidebar is shown in the UI.
	// nil means the default is used (shown).
	ShowAlbumTree *bool
	// ShowHiddenAlbums controls whether personally-hidden albums are shown
	// (dimmed, with a click-to-unhide affordance) instead of being excluded
	// from navigation and search. nil means the default is used (excluded).
	ShowHiddenAlbums *bool
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

	if u.SearchResultLimit != nil && *u.SearchResultLimit < 0 {
		return errors.New("search result limit must not be negative")
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

// EffectiveGrant returns the user's own UserAlbums row on this exact album,
// or nil if they have no access to it. Every album a user can reach always
// has its own row for them (see PropagateAlbumLevel and the album-creation
// call sites that copy a parent's rows onto new children), so a single
// lookup is enough — no ancestor walk needed.
func (user *User) EffectiveGrant(db *gorm.DB, album *Album) (*UserAlbums, error) {
	var grant UserAlbums
	err := db.Where("user_id = ? AND album_id = ?", user.ID, album.ID).First(&grant).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, errors.Wrap(err, "get user's grant on album")
	}
	return &grant, nil
}

// HasAlbumLevel returns true if the user is an admin, or holds at least the
// given AlbumPermissionLevel on album (either directly or because they're
// admin, in which case every level is implicitly satisfied).
func (user *User) HasAlbumLevel(db *gorm.DB, album *Album, level AlbumPermissionLevel) (bool, error) {
	if user.Admin {
		return true, nil
	}

	grant, err := user.EffectiveGrant(db, album)
	if err != nil {
		return false, err
	}

	return grant != nil && grant.Level.AtLeast(level), nil
}

// HideAlbum sets/clears an album as hidden from the user's own navigation.
// This is a personal preference only: it never affects other users'
// visibility of, or access to, the album.
func (user *User) HideAlbum(db *gorm.DB, albumID int, hidden bool) (*Album, error) {
	userAlbumData := UserAlbumData{
		UserID:  user.ID,
		AlbumID: albumID,
		Hidden:  hidden,
	}

	if err := db.Clauses(clause.OnConflict{UpdateAll: true}).Create(&userAlbumData).Error; err != nil {
		return nil, errors.Wrapf(err, "update user hidden album in database")
	}

	var album Album
	if err := db.First(&album, albumID).Error; err != nil {
		return nil, errors.Wrap(err, "get album from database after hide update")
	}

	return &album, nil
}

// UnhideAllAlbums clears the hidden flag on every album this user has
// hidden.
func (user *User) UnhideAllAlbums(db *gorm.DB) error {
	return db.Model(&UserAlbumData{}).
		Where("user_id = ? AND hidden = true", user.ID).
		Update("hidden", false).Error
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
