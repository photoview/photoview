package api

import (
	"context"
	"database/sql"
	"errors"

	"github.com/99designs/gqlgen/graphql"
	"github.com/viktorstrate/photoview/api/graphql/auth"
)

func IsAdmin(database *sql.DB) func(ctx context.Context, obj interface{}, next graphql.Resolver) (res interface{}, err error) {
	return func(ctx context.Context, obj interface{}, next graphql.Resolver) (res interface{}, err error) {

		user := auth.UserFromContext(ctx)
		if user == nil || user.Admin == false {
			return nil, errors.New("user must be admin")
		}

		return next(ctx)
	}
}
