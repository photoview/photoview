package resolvers

import (
	"context"

	"github.com/photoview/photoview/api/graphql/auth"
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func (r *queryResolver) User(ctx context.Context, filter *models.Filter) ([]*models.User, error) {

	var users []*models.User

	if err := filter.FormatSQL(r.Database.Model(models.User{})).Error; err != nil {
		return nil, err
	}

	return users, nil
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
func (r *mutationResolver) RegisterUser(ctx context.Context, username string, password string, rootPath string) (*models.AuthorizeResult, error) {

	var token *models.AccessToken

	transactionError := r.Database.Transaction(func(tx *gorm.DB) error {
		user, err := models.RegisterUser(tx, username, &password, rootPath, false)
		if err != nil {
			return err
		}

		token, err = user.GenerateAccessToken(tx)
		if err != nil {
			tx.Rollback()
			return err
		}

		return nil
	})

	if transactionError != nil {
		return &models.AuthorizeResult{
			Success: false,
			Status:  transactionError.Error(),
		}, transactionError
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

		user, err := models.RegisterUser(tx, username, &password, rootPath, true)
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
func (r *mutationResolver) UpdateUser(ctx context.Context, id int, username *string, rootPath *string, password *string, admin *bool) (*models.User, error) {

	if username == nil && rootPath == nil && password == nil && admin == nil {
		return nil, errors.New("no updates requested")
	}

	var user models.User
	if err := r.Database.First(&user, id).Error; err != nil {
		return nil, err
	}

	if username != nil {
		user.Username = *username
	}

	if rootPath != nil {
		user.RootPath = *rootPath
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

func (r *mutationResolver) CreateUser(ctx context.Context, username string, rootPath string, password *string, admin bool) (*models.User, error) {

	var user *models.User

	transactionError := r.Database.Transaction(func(tx *gorm.DB) error {
		var err error
		user, err = models.RegisterUser(tx, username, password, rootPath, admin)
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
