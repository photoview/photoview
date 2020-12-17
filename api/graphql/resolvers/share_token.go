package resolvers

import (
	"context"
	"database/sql"
	"time"

	"github.com/pkg/errors"

	api "github.com/photoview/photoview/api/graphql"
	"github.com/photoview/photoview/api/graphql/auth"
	"github.com/photoview/photoview/api/graphql/models"
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
	row := r.Database.QueryRow("SELECT * FROM user WHERE user.user_id = ?", obj.OwnerID)
	return models.NewUserFromRow(row)
}

func (r *shareTokenResolver) Album(ctx context.Context, obj *models.ShareToken) (*models.Album, error) {
	row := r.Database.QueryRow("SELECT * FROM album WHERE album.album_id = ?", obj.AlbumID)
	album, err := models.NewAlbumFromRow(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		} else {
			return nil, errors.Wrap(err, "could not get album of share token from database")
		}
	}

	return album, nil
}

func (r *shareTokenResolver) Media(ctx context.Context, obj *models.ShareToken) (*models.Media, error) {
	row := r.Database.QueryRow("SELECT * FROM media WHERE media.media_id = ?", obj.MediaID)
	media, err := models.NewMediaFromRow(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		} else {
			return nil, errors.Wrap(err, "could not get media of share token from database")
		}
	}

	return media, nil
}

func (r *shareTokenResolver) HasPassword(ctx context.Context, obj *models.ShareToken) (bool, error) {
	hasPassword := obj.Password != nil
	return hasPassword, nil
}

func (r *queryResolver) ShareToken(ctx context.Context, tokenValue string, password *string) (*models.ShareToken, error) {

	row := r.Database.QueryRow("SELECT * FROM share_token WHERE value = ?", tokenValue)
	token, err := models.NewShareTokenFromRow(row)
	if err != nil {
		if err == sql.ErrNoRows {
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

	return token, nil
}

func (r *queryResolver) ShareTokenValidatePassword(ctx context.Context, tokenValue string, password *string) (bool, error) {
	row := r.Database.QueryRow("SELECT * FROM share_token WHERE value = ?", tokenValue)
	token, err := models.NewShareTokenFromRow(row)
	if err != nil {
		if err == sql.ErrNoRows {
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

	rows, err := r.Database.Query("SELECT owner_id FROM album WHERE album.album_id = ? AND album.owner_id = ?", albumID, user.UserID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to validate album owner with database")
	}
	if rows.Next() == false {
		return nil, auth.ErrUnauthorized
	}
	rows.Close()

	var hashed_password *string = nil
	if password != nil {
		hashedPassBytes, err := bcrypt.GenerateFromPassword([]byte(*password), 12)
		if err != nil {
			return nil, errors.Wrap(err, "failed to hash token password")
		}
		hashed_str := string(hashedPassBytes)
		hashed_password = &hashed_str
	}

	token := utils.GenerateToken()
	res, err := r.Database.Exec("INSERT INTO share_token (value, owner_id, expire, password, album_id) VALUES (?, ?, ?, ?, ?)", token, user.UserID, expire, hashed_password, albumID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to insert new share token into database")
	}

	token_id, err := res.LastInsertId()
	if err != nil {
		return nil, errors.Wrap(err, "could not get database id of new album share token")
	}

	return &models.ShareToken{
		TokenID:  int(token_id),
		Value:    token,
		OwnerID:  user.UserID,
		Expire:   expire,
		Password: password,
		AlbumID:  &albumID,
		MediaID:  nil,
	}, nil
}

func (r *mutationResolver) ShareMedia(ctx context.Context, mediaID int, expire *time.Time, password *string) (*models.ShareToken, error) {
	user := auth.UserFromContext(ctx)
	if user == nil {
		return nil, auth.ErrUnauthorized
	}

	rows, err := r.Database.Query("SELECT owner_id FROM album, media WHERE media.media_id = ? AND media.album_id = album.album_id AND album.owner_id = ?", mediaID, user.UserID)
	if err != nil {
		return nil, errors.Wrap(err, "error validating owner of media with database")
	}
	if rows.Next() == false {
		return nil, auth.ErrUnauthorized
	}
	rows.Close()

	hashed_password, err := hashSharePassword(password)
	if err != nil {
		return nil, err
	}

	token := utils.GenerateToken()
	res, err := r.Database.Exec("INSERT INTO share_token (value, owner_id, expire, password, media_id) VALUES (?, ?, ?, ?, ?)", token, user.UserID, expire, hashed_password, mediaID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to insert new share token into database")
	}

	token_id, err := res.LastInsertId()
	if err != nil {
		return nil, errors.Wrap(err, "could not get database id of new media share token")
	}

	return &models.ShareToken{
		TokenID:  int(token_id),
		Value:    token,
		OwnerID:  user.UserID,
		Expire:   expire,
		Password: password,
		AlbumID:  nil,
		MediaID:  &mediaID,
	}, nil
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

	if _, err := r.Database.Exec("DELETE FROM share_token WHERE token_id = ?", token.TokenID); err != nil {
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

	hashed_password, err := hashSharePassword(password)
	if err != nil {
		return nil, err
	}

	_, err = r.Database.Exec("UPDATE share_token SET password = ? WHERE token_id = ?", hashed_password, token.TokenID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to update password for share token")
	}

	updatedToken := r.Database.QueryRow("SELECT * FROM share_token WHERE value = ?", tokenValue)
	return models.NewShareTokenFromRow(updatedToken)
}

func hashSharePassword(password *string) (*string, error) {
	var hashed_password *string = nil
	if password != nil {
		hashedPassBytes, err := bcrypt.GenerateFromPassword([]byte(*password), 12)
		if err != nil {
			return nil, errors.Wrap(err, "failed to generate hash for share password")
		}
		hashed_str := string(hashedPassBytes)
		hashed_password = &hashed_str
	}

	return hashed_password, nil
}

func getUserToken(db *sql.DB, user *models.User, tokenValue string) (*models.ShareToken, error) {
	row := db.QueryRow(`
		SELECT share_token.* FROM share_token, user WHERE
		share_token.value = ? AND
		share_token.owner_id = user.user_id AND
		(user.user_id = ? OR user.admin = TRUE)
	`, tokenValue, user.UserID)

	token, err := models.NewShareTokenFromRow(row)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get user share token from database")
	}

	return token, nil
}
