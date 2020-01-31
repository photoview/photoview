package api

import (
	"context"
	"fmt"

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

	token := fmt.Sprintf("token:%d", user.User_id)

	return &AuthorizeResult{
		Success: true,
		Status:  "ok",
		Token:   &token,
	}, nil
}
func (r *mutationResolver) RegisterUser(ctx context.Context, username string, password string) (*AuthorizeResult, error) {
	user, err := models.RegisterUser(r.Database, username, password)
	if err != nil {
		return &AuthorizeResult{
			Success: false,
			Status:  err.Error(),
		}, nil
	}

	token := fmt.Sprintf("token:%d", user.User_id)

	return &AuthorizeResult{
		Success: true,
		Status:  "ok",
		Token:   &token,
	}, nil
}
