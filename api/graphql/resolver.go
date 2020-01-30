package api

import (
	"context"
) // THIS CODE IS A STARTING POINT ONLY. IT WILL NOT BE UPDATED WITH SCHEMA CHANGES.

type Resolver struct{}

func (r *Resolver) Mutation() MutationResolver {
	return &mutationResolver{r}
}
func (r *Resolver) Query() QueryResolver {
	return &queryResolver{r}
}

type mutationResolver struct{ *Resolver }

func (r *mutationResolver) AuthorizeUser(ctx context.Context, username string, password string) (*AuthorizeResult, error) {
	panic("not implemented")
}
func (r *mutationResolver) RegisterUser(ctx context.Context, username string, password string) (*AuthorizeResult, error) {
	panic("not implemented")
}

type queryResolver struct{ *Resolver }

func (r *queryResolver) Users(ctx context.Context) ([]*User, error) {
	users := make([]*User, 0)

	return users, nil
}
