package api

import (
	"context"
	"database/sql"

	"github.com/viktorstrate/photoview/api/graphql/models"
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

func (r *queryResolver) Users(ctx context.Context) ([]*models.User, error) {

	rows, err := r.Database.Query("SELECT * FROM users")
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

func (r *queryResolver) MyAlbums(ctx context.Context) ([]*models.Album, error) {
	panic("Not implemented")
}
func (r *queryResolver) Album(ctx context.Context, id *string) (*models.Album, error) {
	panic("Not implemented")
}
func (r *queryResolver) MyPhotos(ctx context.Context) ([]*models.Photo, error) {
	panic("Not implemented")
}
func (r *queryResolver) Photo(ctx context.Context, id string) (*models.Photo, error) {
	panic("Not implemented")
}
