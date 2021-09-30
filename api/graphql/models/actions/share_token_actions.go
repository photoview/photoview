package actions

import (
	"time"

	"github.com/photoview/photoview/api/graphql/auth"
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/utils"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func AddMediaShare(db *gorm.DB, user *models.User, mediaID int, expire *time.Time, password *string) (*models.ShareToken, error) {
	var media models.Media

	var query string
	if db.Dialector.Name() == "postgres" {
		query = "EXISTS (SELECT * FROM user_albums WHERE user_albums.album_id = \"Album\".id AND user_albums.user_id = ?)"
	} else {
		query = "EXISTS (SELECT * FROM user_albums WHERE user_albums.album_id = Album.id AND user_albums.user_id = ?)"
	}

	err := db.Joins("Album").
		Where(query, user.ID).
		First(&media, mediaID).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, auth.ErrUnauthorized
		} else {
			return nil, errors.Wrap(err, "failed to validate media owner with database")
		}
	}

	hashedPassword, err := hashSharePassword(password)
	if err != nil {
		return nil, err
	}

	shareToken := models.ShareToken{
		Value:    utils.GenerateToken(),
		OwnerID:  user.ID,
		Expire:   expire,
		Password: hashedPassword,
		AlbumID:  nil,
		MediaID:  &mediaID,
	}

	if err := db.Create(&shareToken).Error; err != nil {
		return nil, errors.Wrap(err, "failed to insert new share token into database")
	}

	return &shareToken, nil
}

func AddAlbumShare(db *gorm.DB, user *models.User, albumID int, expire *time.Time, password *string) (*models.ShareToken, error) {
	var count int64
	err := db.
		Model(&models.Album{}).
		Where("EXISTS (SELECT * FROM user_albums WHERE user_albums.album_id = albums.id AND user_albums.user_id = ?)", user.ID).
		Count(&count).Error

	if err != nil {
		return nil, errors.Wrap(err, "failed to validate album owner with database")
	}

	if count == 0 {
		return nil, auth.ErrUnauthorized
	}

	var hashedPassword *string = nil
	if password != nil {
		hashedPassBytes, err := bcrypt.GenerateFromPassword([]byte(*password), 12)
		if err != nil {
			return nil, errors.Wrap(err, "failed to hash token password")
		}
		hashedStr := string(hashedPassBytes)
		hashedPassword = &hashedStr
	}

	shareToken := models.ShareToken{
		Value:    utils.GenerateToken(),
		OwnerID:  user.ID,
		Expire:   expire,
		Password: hashedPassword,
		AlbumID:  &albumID,
		MediaID:  nil,
	}

	if err := db.Create(&shareToken).Error; err != nil {
		return nil, errors.Wrap(err, "failed to insert new share token into database")
	}

	return &shareToken, nil
}

func DeleteShareToken(db *gorm.DB, userID int, tokenValue string) (*models.ShareToken, error) {
	token, err := getUserToken(db, userID, tokenValue)
	if err != nil {
		return nil, err
	}

	if err := db.Delete(&token).Error; err != nil {
		return nil, errors.Wrapf(err, "failed to delete share token (%s) from database", tokenValue)
	}

	return token, nil
}

func ProtectShareToken(db *gorm.DB, userID int, tokenValue string, password *string) (*models.ShareToken, error) {
	token, err := getUserToken(db, userID, tokenValue)
	if err != nil {
		return nil, err
	}

	hashedPassword, err := hashSharePassword(password)
	if err != nil {
		return nil, err
	}

	token.Password = hashedPassword

	if err := db.Save(&token).Error; err != nil {
		return nil, errors.Wrap(err, "failed to update password for share token")
	}

	return token, nil
}

func hashSharePassword(password *string) (*string, error) {
	var hashedPassword *string = nil
	if password != nil {
		hashedPassBytes, err := bcrypt.GenerateFromPassword([]byte(*password), 12)
		if err != nil {
			return nil, errors.Wrap(err, "failed to generate hash for share password")
		}
		hashedStr := string(hashedPassBytes)
		hashedPassword = &hashedStr
	}

	return hashedPassword, nil
}

func getUserToken(db *gorm.DB, userID int, tokenValue string) (*models.ShareToken, error) {

	var query string
	if db.Dialector.Name() == "postgres" {
		query = "\"Owner\".id = ? OR \"Owner\".admin = TRUE"
	} else {
		query = "Owner.id = ? OR Owner.admin = TRUE"
	}

	var token models.ShareToken
	err := db.Where("share_tokens.value = ?", tokenValue).Joins("Owner").Where(query, userID).First(&token).Error

	if err != nil {
		return nil, errors.Wrap(err, "failed to get user share token from database")
	}

	return &token, nil
}
