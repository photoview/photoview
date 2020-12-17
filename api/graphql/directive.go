package api

import (
	"context"
	"errors"

	"github.com/99designs/gqlgen/graphql"
	"github.com/photoview/photoview/api/graphql/auth"
	"gorm.io/gorm"
)

func IsAdmin(database *gorm.DB) func(ctx context.Context, obj interface{}, next graphql.Resolver) (res interface{}, err error) {
	return func(ctx context.Context, obj interface{}, next graphql.Resolver) (res interface{}, err error) {

		user := auth.UserFromContext(ctx)
		if user == nil || user.Admin == false {
			return nil, errors.New("user must be admin")
		}

		return next(ctx)
	}
}
