package resolvers

import (
	"context"
	"database/sql"

	api "github.com/viktorstrate/photoview/api/graphql"
	"github.com/viktorstrate/photoview/api/graphql/models"
)

//go:generate go run github.com/99designs/gqlgen

type Resolver struct {
	Database *sql.DB
}

func (r *Resolver) Mutation() api.MutationResolver {
	return &mutationResolver{r}
}

func (r *Resolver) Query() api.QueryResolver {
	return &queryResolver{r}
}

type mutationResolver struct{ *Resolver }

type queryResolver struct{ *Resolver }

func (r *queryResolver) SiteInfo(ctx context.Context) (*models.SiteInfo, error) {
	return models.GetSiteInfo(r.Database)
}
