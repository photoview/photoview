package resolvers

import (
	"context"

	api "github.com/viktorstrate/photoview/api/graphql"
	"github.com/viktorstrate/photoview/api/graphql/models"
	"gorm.io/gorm"
)

//go:generate go run github.com/99designs/gqlgen

type Resolver struct {
	Database *gorm.DB
}

func (r *Resolver) Mutation() api.MutationResolver {
	return &mutationResolver{r}
}

func (r *Resolver) Query() api.QueryResolver {
	return &queryResolver{r}
}

func (r *Resolver) Subscription() api.SubscriptionResolver {
	return &subscriptionResolver{
		Resolver: r,
	}
}

type mutationResolver struct{ *Resolver }

type queryResolver struct{ *Resolver }

type subscriptionResolver struct {
	Resolver *Resolver
}

func (r *queryResolver) SiteInfo(ctx context.Context) (*models.SiteInfo, error) {
	return models.GetSiteInfo(r.Database)
}
