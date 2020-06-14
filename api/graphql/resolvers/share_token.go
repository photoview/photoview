package resolvers

import (
	"context"
	"database/sql"
	"time"

	"github.com/pkg/errors"

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
			return nil, err
		}
	}

	return album, nil
}

func (r *shareTokenResolver) Photo(ctx context.Context, obj *models.ShareToken) (*models.Photo, error) {
	row := r.Database.QueryRow("SELECT * FROM photo WHERE photo.photo_id = ?", obj.PhotoID)
	photo, err := models.NewPhotoFromRow(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		} else {
			return nil, err
		}
	}

	return photo, nil
}

func (r *shareTokenResolver) HasPassword(ctx context.Context, obj *models.ShareToken) (bool, error) {
	hasPassword := obj.Password != nil
	return hasPassword, nil
}

func (r *queryResolver) ShareToken(ctx context.Context, token string, password *string) (*models.ShareToken, error) {

	row := r.Database.QueryRow("SELECT * FROM share_token WHERE value = ? AND (password = ? OR password IS NULL)", token, password)
	result, err := models.NewShareTokenFromRow(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("share not found")
		} else {
			return nil, err
		}
	}

	return result, nil
}

func (r *mutationResolver) ShareAlbum(ctx context.Context, albumID int, expire *time.Time, password *string) (*models.ShareToken, error) {
	user := auth.UserFromContext(ctx)
	if user == nil {
		return nil, auth.ErrUnauthorized
	}

	rows, err := r.Database.Query("SELECT owner_id FROM album WHERE album.album_id = ? AND album.owner_id = ?", albumID, user.UserID)
	if err != nil {
		return nil, err
	}
	if rows.Next() == false {
		return nil, auth.ErrUnauthorized
	}
	rows.Close()

	var hashed_password *string = nil
	if password != nil {
		hashedPassBytes, err := bcrypt.GenerateFromPassword([]byte(*password), 12)
		if err != nil {
			return nil, err
		}
		hashed_str := string(hashedPassBytes)
		hashed_password = &hashed_str
	}

	token := utils.GenerateToken()
	res, err := r.Database.Exec("INSERT INTO share_token (value, owner_id, expire, password, album_id) VALUES (?, ?, ?, ?, ?)", token, user.UserID, expire, hashed_password, albumID)
	if err != nil {
		return nil, err
	}

	token_id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &models.ShareToken{
		TokenID:  int(token_id),
		Value:    token,
		OwnerID:  user.UserID,
		Expire:   expire,
		Password: password,
		AlbumID:  &albumID,
		PhotoID:  nil,
	}, nil
}

func (r *mutationResolver) SharePhoto(ctx context.Context, photoID int, expire *time.Time, password *string) (*models.ShareToken, error) {
	user := auth.UserFromContext(ctx)
	if user == nil {
		return nil, auth.ErrUnauthorized
	}

	rows, err := r.Database.Query("SELECT owner_id FROM album, photo WHERE photo.photo_id = ? AND photo.album_id = album.album_id AND album.owner_id = ?", photoID, user.UserID)
	if err != nil {
		return nil, err
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
	res, err := r.Database.Exec("INSERT INTO share_token (value, owner_id, expire, password, photo_id) VALUES (?, ?, ?, ?, ?)", token, user.UserID, expire, hashed_password, photoID)
	if err != nil {
		return nil, err
	}

	token_id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &models.ShareToken{
		TokenID:  int(token_id),
		Value:    token,
		OwnerID:  user.UserID,
		Expire:   expire,
		Password: password,
		AlbumID:  nil,
		PhotoID:  &photoID,
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
		return nil, errors.Wrapf(err, "Error occurred when trying to delete share token (%s) from database", tokenValue)
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
		return nil, errors.Wrap(err, "Failed to update password for share token")
	}

	updatedToken := r.Database.QueryRow("SELECT * FROM share_token WHERE value = ?", tokenValue)
	return models.NewShareTokenFromRow(updatedToken)
}

func hashSharePassword(password *string) (*string, error) {
	var hashed_password *string = nil
	if password != nil {
		hashedPassBytes, err := bcrypt.GenerateFromPassword([]byte(*password), 12)
		if err != nil {
			return nil, err
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
		return nil, err
	}

	return token, nil
}
