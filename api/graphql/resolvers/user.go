package resolvers

import (
	"context"
	"errors"
	"log"

	"github.com/viktorstrate/photoview/api/graphql/auth"
	"github.com/viktorstrate/photoview/api/graphql/models"
	"golang.org/x/crypto/bcrypt"
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

	user, err := models.RegisterUser(tx, username, &password, rootPath, false)
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

	user, err := models.RegisterUser(tx, username, &password, rootPath, true)
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

// Admin queries
func (r *mutationResolver) UpdateUser(ctx context.Context, id int, username *string, rootPath *string, password *string, admin *bool) (*models.User, error) {

	user_rows, err := r.Database.Query("SELECT * FROM user WHERE user_id = ?", id)
	if err != nil {
		return nil, err
	}
	if user_rows.Next() == false {
		return nil, errors.New("user not found")
	}
	user_rows.Close()

	update_str := ""
	update_args := make([]interface{}, 0)

	if username != nil {
		update_str += "username = ?, "
		update_args = append(update_args, username)
	}
	if rootPath != nil {
		if !models.ValidRootPath(*rootPath) {
			return nil, errors.New("invalid root path")
		}

		update_str += "root_path = ?, "
		update_args = append(update_args, rootPath)
	}
	if admin != nil {
		update_str += "admin = ?, "
		update_args = append(update_args, admin)
	}
	if password != nil {
		hashedPassBytes, err := bcrypt.GenerateFromPassword([]byte(*password), 12)
		if err != nil {
			return nil, err
		}
		hashedPass := string(hashedPassBytes)

		update_str += "password = ?, "
		update_args = append(update_args, hashedPass)
	}

	if len(update_str) == 0 {
		return nil, errors.New("no updates requested")
	}

	update_str = update_str[:len(update_str)-2]
	log.Printf("Updating user with update string: %s\n", update_str)

	update_args = append(update_args, id)

	res, err := r.Database.Exec("UPDATE user SET "+update_str+" WHERE user_id = ?", update_args...)
	if err != nil {
		return nil, err
	}
	rows_aff, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}
	if rows_aff == 0 {
		return nil, errors.New("no users were updated")
	}

	row := r.Database.QueryRow("SELECT * FROM user WHERE user_id = ?", id)
	user, err := models.NewUserFromRow(row)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *mutationResolver) CreateUser(ctx context.Context, username string, rootPath string, password *string, admin bool) (*models.User, error) {
	tx, err := r.Database.Begin()
	if err != nil {
		return nil, err
	}

	user, err := models.RegisterUser(tx, username, password, rootPath, admin)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return user, nil
}

func (r *mutationResolver) DeleteUser(ctx context.Context, id int) (*models.User, error) {

	row := r.Database.QueryRow("SELECT * FROM user WHERE user_id = ?", id)
	user, err := models.NewUserFromRow(row)
	if err != nil {
		return nil, err
	}

	res, err := r.Database.Exec("DELETE FROM user WHERE user_id = ?", id)
	if err != nil {
		return nil, err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}
	if rows == 0 {
		return nil, errors.New("no users deleted")
	}

	return user, nil
}
