package resolvers

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	api "github.com/photoview/photoview/api/graphql"
	"github.com/photoview/photoview/api/graphql/auth"
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/graphql/models/actions"
	"github.com/photoview/photoview/api/utils"
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

func (r *queryResolver) ShareToken(ctx context.Context, credentials models.ShareTokenCredentials) (*models.ShareToken, error) {

	var token models.ShareToken
	if err := r.Database.Preload(clause.Associations).Where("value = ?", credentials.Token).First(&token).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("share not found")
		} else {
			return nil, errors.Wrap(err, "failed to get share token from database")
		}
	}

	if token.Password != nil {
		if err := bcrypt.CompareHashAndPassword([]byte(*token.Password), []byte(*credentials.Password)); err != nil {
			if err == bcrypt.ErrMismatchedHashAndPassword {
				return nil, errors.New("unauthorized")
			} else {
				return nil, errors.Wrap(err, "failed to compare token password hashes")
			}
		}
	}

	return &token, nil
}

func (r *queryResolver) ShareTokenValidatePassword(ctx context.Context, credentials models.ShareTokenCredentials) (bool, error) {
	var token models.ShareToken
	if err := r.Database.Where("value = ?", credentials.Token).First(&token).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, errors.New("share not found")
		} else {
			return false, errors.Wrap(err, "failed to get share token from database")
		}
	}

	if token.Password == nil {
		return true, nil
	}

	if credentials.Password == nil {
		return false, nil
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*token.Password), []byte(*credentials.Password)); err != nil {
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
	err := r.Database.
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

	return actions.AddMediaShare(r.Database, user.ID, mediaID, expire, password)
}

func (r *mutationResolver) DeleteShareToken(ctx context.Context, tokenValue string) (*models.ShareToken, error) {
	user := auth.UserFromContext(ctx)
	if user == nil {
		return nil, auth.ErrUnauthorized
	}

	return actions.DeleteShareToken(r.Database, user.ID, tokenValue)
}

func (r *mutationResolver) ProtectShareToken(ctx context.Context, tokenValue string, password *string) (*models.ShareToken, error) {
	user := auth.UserFromContext(ctx)
	if user == nil {
		return nil, auth.ErrUnauthorized
	}

	return actions.ProtectShareToken(r.Database, user.ID, tokenValue, password)
}
