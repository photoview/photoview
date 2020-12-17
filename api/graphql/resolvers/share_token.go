package resolvers

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	api "github.com/viktorstrate/photoview/api/graphql"
	"github.com/viktorstrate/photoview/api/graphql/auth"
	"github.com/viktorstrate/photoview/api/graphql/models"
	"github.com/viktorstrate/photoview/api/utils"
	"golang.org/x/crypto/bcrypt"
)

type shareTokenResolver struct {
	*Resolver
}

func (r *Resolver) ShareToken() api.ShareTokenResolver {
	return &shareTokenResolver{r}
}

func (r *shareTokenResolver) Owner(ctx context.Context, obj *models.ShareToken) (*models.User, error) {
	return &obj.Owner, nil
}

func (r *shareTokenResolver) Album(ctx context.Context, obj *models.ShareToken) (*models.Album, error) {
	return obj.Album, nil
}

func (r *shareTokenResolver) Media(ctx context.Context, obj *models.ShareToken) (*models.Media, error) {
	return obj.Media, nil
}

func (r *shareTokenResolver) HasPassword(ctx context.Context, obj *models.ShareToken) (bool, error) {
	hasPassword := obj.Password != nil
	return hasPassword, nil
}

func (r *queryResolver) ShareToken(ctx context.Context, tokenValue string, password *string) (*models.ShareToken, error) {

	var token models.ShareToken
	if err := r.Database.Preload(clause.Associations).Where("value = ?", tokenValue).First(&token).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("share not found")
		} else {
			return nil, errors.Wrap(err, "failed to get share token from database")
		}
	}

	if token.Password != nil {
		if err := bcrypt.CompareHashAndPassword([]byte(*token.Password), []byte(*password)); err != nil {
			if err == bcrypt.ErrMismatchedHashAndPassword {
				return nil, errors.New("unauthorized")
			} else {
				return nil, errors.Wrap(err, "failed to compare token password hashes")
			}
		}
	}

	return &token, nil
}

func (r *queryResolver) ShareTokenValidatePassword(ctx context.Context, tokenValue string, password *string) (bool, error) {
	var token models.ShareToken
	if err := r.Database.Where("value = ?", tokenValue).First(&token).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, errors.New("share not found")
		} else {
			return false, errors.Wrap(err, "failed to get share token from database")
		}
	}

	if token.Password == nil {
		return true, nil
	}

	if password == nil {
		return false, nil
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*token.Password), []byte(*password)); err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return false, nil
		} else {
			return false, errors.Wrap(err, "could not compare token password hashes")
		}
	}

	return true, nil
}

func (r *mutationResolver) ShareAlbum(ctx context.Context, albumID int, expire *time.Time, password *string) (*models.ShareToken, error) {
	user := auth.UserFromContext(ctx)
	if user == nil {
		return nil, auth.ErrUnauthorized
	}

	var count int64
	if err := r.Database.Model(&models.Album{}).Where("owner_id = ?", user.ID).Count(&count).Error; err != nil {
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

	if err := r.Database.Create(&shareToken).Error; err != nil {
		return nil, errors.Wrap(err, "failed to insert new share token into database")
	}

	return &shareToken, nil
}

func (r *mutationResolver) ShareMedia(ctx context.Context, mediaID int, expire *time.Time, password *string) (*models.ShareToken, error) {
	user := auth.UserFromContext(ctx)
	if user == nil {
		return nil, auth.ErrUnauthorized
	}

	var media models.Media

	if err := r.Database.Joins("Album").Where("Album.owner_id = ?", user.ID).First(&media, mediaID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, auth.ErrUnauthorized
		} else {
			return nil, errors.Wrap(err, "failed to validate media owner with database")
		}
	}

	var count int64

	err := r.Database.Raw("SELECT owner_id FROM albums, media WHERE media.id = ? AND media.album_id = albums.id AND albums.owner_id = ?", mediaID, user.ID).Count(&count).Error
	if err != nil {
		return nil, errors.Wrap(err, "error validating owner of media with database")
	}

	if count == 0 {
		return nil, auth.ErrUnauthorized
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

	if err := r.Database.Create(&shareToken).Error; err != nil {
		return nil, errors.Wrap(err, "failed to insert new share token into database")
	}

	return &shareToken, nil
}

func (r *mutationResolver) DeleteShareToken(ctx context.Context, tokenValue string) (*models.ShareToken, error) {
	user := auth.UserFromContext(ctx)
	if user == nil {
		return nil, auth.ErrUnauthorized
	}

	token, err := getUserToken(r.Database, user, tokenValue)
	if err != nil {
		return nil, err
	}

	if err := r.Database.Delete(&token).Error; err != nil {
		return nil, errors.Wrapf(err, "failed to delete share token (%s) from database", tokenValue)
	}

	return token, nil
}

func (r *mutationResolver) ProtectShareToken(ctx context.Context, tokenValue string, password *string) (*models.ShareToken, error) {
	user := auth.UserFromContext(ctx)
	if user == nil {
		return nil, auth.ErrUnauthorized
	}

	token, err := getUserToken(r.Database, user, tokenValue)
	if err != nil {
		return nil, err
	}

	hashedPassword, err := hashSharePassword(password)
	if err != nil {
		return nil, err
	}

	token.Password = hashedPassword

	if err := r.Database.Save(&token).Error; err != nil {
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

func getUserToken(db *gorm.DB, user *models.User, tokenValue string) (*models.ShareToken, error) {

	var token models.ShareToken
	err := db.Where("share_tokens.value = ?", tokenValue).Joins("Owner").Where("Owner.id = ? OR Owner.admin = TRUE", user.ID).First(&token).Error

	if err != nil {
		return nil, errors.Wrap(err, "failed to get user share token from database")
	}

	return &token, nil
}
