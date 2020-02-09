package resolvers

import (
	"context"
	"database/sql"
	"log"
	"time"

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

func (r *queryResolver) AlbumShares(ctx context.Context, id int, password *string) ([]*models.ShareToken, error) {
	log.Println("Query AlbumShares: not implemented")

	tokens := make([]*models.ShareToken, 0)
	return tokens, nil
}

func (r *queryResolver) PhotoShares(ctx context.Context, id int, password *string) ([]*models.ShareToken, error) {
	log.Println("Query PhotoShares: not implemented")

	tokens := make([]*models.ShareToken, 0)
	return tokens, nil
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
	panic("not implemented")
}
