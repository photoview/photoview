package resolvers

import (
	"context"

	api "github.com/photoview/photoview/api/graphql"
	"github.com/photoview/photoview/api/graphql/auth"
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type userResolver struct {
	*Resolver
}

func (r *Resolver) User() api.UserResolver {
	return &userResolver{r}
}

func (r *queryResolver) User(ctx context.Context, filter *models.Filter) ([]*models.User, error) {

	var users []*models.User

	if err := filter.FormatSQL(r.Database.Model(models.User{})).Scan(&users).Error; err != nil {
		return nil, err
	}

	return users, nil
}

func (r *userResolver) Albums(ctx context.Context, user *models.User) ([]*models.Album, error) {
	user.FillAlbums(r.Database)

	pointerAlbums := make([]*models.Album, len(user.Albums))
	for i, album := range user.Albums {
		pointerAlbums[i] = &album
	}

	return pointerAlbums, nil
}

func (r *userResolver) RootAlbums(ctx context.Context, user *models.User) (albums []*models.Album, err error) {

	err = r.Database.Model(&user).
		Where("albums.parent_album_id NOT IN (?)",
			r.Database.Table("user_albums").
				Select("albums.id").
				Joins("JOIN albums ON albums.id = user_albums.album_id AND user_albums.user_id = ?", user.ID),
		).Or("albums.parent_album_id IS NULL").
		Association("Albums").Find(&albums)

	return
}

func (r *queryResolver) MyUser(ctx context.Context) (*models.User, error) {

	user := auth.UserFromContext(ctx)
	if user == nil {
		return nil, auth.ErrUnauthorized
	}

	return user, nil
}

func (r *mutationResolver) AuthorizeUser(ctx context.Context, username string, password string) (*models.AuthorizeResult, error) {
	user, err := models.AuthorizeUser(r.Database, username, password)
	if err != nil {
		return &models.AuthorizeResult{
			Success: false,
			Status:  err.Error(),
		}, nil
	}

	var token *models.AccessToken

	transactionError := r.Database.Transaction(func(tx *gorm.DB) error {
		token, err = user.GenerateAccessToken(tx)
		if err != nil {
			return err
		}

		return nil
	})

	if transactionError != nil {
		return nil, transactionError
	}

	return &models.AuthorizeResult{
		Success: true,
		Status:  "ok",
		Token:   &token.Value,
	}, nil
}

func (r *mutationResolver) InitialSetupWizard(ctx context.Context, username string, password string, rootPath string) (*models.AuthorizeResult, error) {
	siteInfo, err := models.GetSiteInfo(r.Database)
	if err != nil {
		return nil, err
	}

	if !siteInfo.InitialSetup {
		return nil, errors.New("not initial setup")
	}

	var token *models.AccessToken

	transactionError := r.Database.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("UPDATE site_info SET initial_setup = false").Error; err != nil {
			return err
		}

		user, err := models.RegisterUser(tx, username, &password, true)
		if err != nil {
			return err
		}

		_, err = scanner.NewRootAlbum(tx, rootPath, user)
		if err != nil {
			return err
		}

		token, err = user.GenerateAccessToken(tx)
		if err != nil {
			return err
		}

		return nil
	})

	if transactionError != nil {
		return &models.AuthorizeResult{
			Success: false,
			Status:  err.Error(),
		}, nil
	}

	return &models.AuthorizeResult{
		Success: true,
		Status:  "ok",
		Token:   &token.Value,
	}, nil
}

// Admin queries
func (r *mutationResolver) UpdateUser(ctx context.Context, id int, username *string, password *string, admin *bool) (*models.User, error) {

	if username == nil && password == nil && admin == nil {
		return nil, errors.New("no updates requested")
	}

	var user models.User
	if err := r.Database.First(&user, id).Error; err != nil {
		return nil, err
	}

	if username != nil {
		user.Username = *username
	}

	if password != nil {
		hashedPassBytes, err := bcrypt.GenerateFromPassword([]byte(*password), 12)
		if err != nil {
			return nil, err
		}
		hashedPass := string(hashedPassBytes)

		user.Password = &hashedPass
	}

	if admin != nil {
		user.Admin = *admin
	}

	if err := r.Database.Save(&user).Error; err != nil {
		return nil, errors.Wrap(err, "failed to update user")
	}

	return &user, nil
}

func (r *mutationResolver) CreateUser(ctx context.Context, username string, password *string, admin bool) (*models.User, error) {

	var user *models.User

	transactionError := r.Database.Transaction(func(tx *gorm.DB) error {
		var err error
		user, err = models.RegisterUser(tx, username, password, admin)
		if err != nil {
			return err
		}

		return nil
	})

	if transactionError != nil {
		return nil, transactionError
	}

	return user, nil
}

func (r *mutationResolver) DeleteUser(ctx context.Context, id int) (*models.User, error) {

	var user models.User

	if err := r.Database.First(&user, id).Error; err != nil {
		return nil, err
	}

	if err := r.Database.Delete(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *mutationResolver) UserAddRootPath(ctx context.Context, id int, rootPath string) (*models.Album, error) {

	var user models.User
	if err := r.Database.First(&user, id).Error; err != nil {
		return nil, err
	}

	// TODO: Check if path exists and that user does not already own rootPath, directly or indirectly

	newAlbum, err := scanner.NewRootAlbum(r.Database, rootPath, &user)
	if err != nil {
		return nil, err
	}

	return newAlbum, nil
}

func (r *mutationResolver) UserRemoveRootAlbum(ctx context.Context, userID int, albumID int) (*models.Album, error) {

	var album models.Album
	if err := r.Database.First(&album, albumID).Error; err != nil {
		return nil, err
	}

	result := r.Database.Exec("DELETE FROM user_albums WHERE album_id = ? AND user_id = ?", albumID, userID)
	if result.Error != nil {
		return nil, result.Error
	}

	if result.RowsAffected == 0 {
		return nil, errors.New("No relation deleted")
	}

	return &album, nil
}
