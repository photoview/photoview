package api

import (
	"context"
	"database/sql"
)

//go:generate go run github.com/99designs/gqlgen

type Resolver struct {
	Database *sql.DB
}

func (r *Resolver) Mutation() MutationResolver {
	return &mutationResolver{r}
}
func (r *Resolver) Query() QueryResolver {
	return &queryResolver{r}
}

type mutationResolver struct{ *Resolver }

type queryResolver struct{ *Resolver }

func (r *queryResolver) Users(ctx context.Context) ([]*User, error) {
	users := make([]*User, 0)

	return users, nil
}
