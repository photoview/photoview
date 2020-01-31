package api

import (
	"context"

	"github.com/viktorstrate/photoview/api/graphql/models"
)

func (r *mutationResolver) AuthorizeUser(ctx context.Context, username string, password string) (*AuthorizeResult, error) {
	user, err := models.AuthorizeUser(r.Database, username, password)
	if err != nil {
		return &AuthorizeResult{
			Success: false,
			Status:  err.Error(),
		}, nil
	}

	token, err := user.GenerateAccessToken(r.Database)
	if err != nil {
		return nil, err
	}

	return &AuthorizeResult{
		Success: true,
		Status:  "ok",
		Token:   &token.Value,
	}, nil
}
func (r *mutationResolver) RegisterUser(ctx context.Context, username string, password string, rootPath string) (*AuthorizeResult, error) {
	user, err := models.RegisterUser(r.Database, username, password, rootPath)
	if err != nil {
		return &AuthorizeResult{
			Success: false,
			Status:  err.Error(),
		}, nil
	}

	token, err := user.GenerateAccessToken(r.Database)
	if err != nil {
		return nil, err
	}

	return &AuthorizeResult{
		Success: true,
		Status:  "ok",
		Token:   &token.Value,
	}, nil
}
