package resolvers

import (
	"context"

	api "github.com/photoview/photoview/api/graphql"
	"gorm.io/gorm"
)

//go:generate go run github.com/99designs/gqlgen

type Resolver struct {
	database *gorm.DB
}

func NewRootResolver(db *gorm.DB) Resolver {
	return Resolver{
		database: db,
	}
}

// DB returns a database instance that is tied to the given context
func (r *Resolver) DB(ctx context.Context) *gorm.DB {
	return r.database.WithContext(ctx)
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
