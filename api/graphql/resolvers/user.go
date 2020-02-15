package resolvers

import (
	"context"
	"errors"

	"github.com/viktorstrate/photoview/api/graphql/auth"
	"github.com/viktorstrate/photoview/api/graphql/models"
)

// func (r *Resolver) User() UserResolver {
// 	return &userResolver{r}
// }

// type userResolver struct{ *Resolver }

func (r *queryResolver) User(ctx context.Context, filter *models.Filter) ([]*models.User, error) {

	filterSQL, err := filter.FormatSQL()
	if err != nil {
		return nil, err
	}

	rows, err := r.Database.Query("SELECT * FROM user" + filterSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users, err := models.NewUsersFromRows(rows)
	if err != nil {
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

	tx, err := r.Database.Begin()
	if err != nil {
		return nil, err
	}

	var token *models.AccessToken

	token, err = user.GenerateAccessToken(tx)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	tx.Commit()

	return &models.AuthorizeResult{
		Success: true,
		Status:  "ok",
		Token:   &token.Value,
	}, nil
}
func (r *mutationResolver) RegisterUser(ctx context.Context, username string, password string, rootPath string) (*models.AuthorizeResult, error) {
	tx, err := r.Database.Begin()
	if err != nil {
		return nil, err
	}

	user, err := models.RegisterUser(tx, username, password, rootPath)
	if err != nil {
		tx.Rollback()
		return &models.AuthorizeResult{
			Success: false,
			Status:  err.Error(),
		}, nil
	}

	token, err := user.GenerateAccessToken(tx)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
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

	tx, err := r.Database.Begin()
	if err != nil {
		return nil, err
	}

	if _, err := tx.Exec("UPDATE site_info SET initial_setup = false"); err != nil {
		tx.Rollback()
		return nil, err
	}

	user, err := models.RegisterUser(tx, username, password, rootPath)
	if err != nil {
		tx.Rollback()
		return &models.AuthorizeResult{
			Success: false,
			Status:  err.Error(),
		}, nil
	}

	if _, err := tx.Exec("UPDATE user SET admin = true WHERE user_id = ?", user.UserID); err != nil {
		tx.Rollback()
		return nil, err
	}

	token, err := user.GenerateAccessToken(tx)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &models.AuthorizeResult{
		Success: true,
		Status:  "ok",
		Token:   &token.Value,
	}, nil
}
