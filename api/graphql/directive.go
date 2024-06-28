package api

import (
	"context"
	"errors"
	"github.com/photoview/photoview/api/graphql/models"

	"github.com/99designs/gqlgen/graphql"
	"github.com/photoview/photoview/api/graphql/auth"
)

func userHasAdminRole(user *models.User) bool {
	return user != nil && user.Role.Name == "ADMIN"

}

func IsAuthorized(ctx context.Context, obj interface{}, next graphql.Resolver) (res interface{}, err error) {
	user := auth.UserFromContext(ctx)
	if user == nil {
		return nil, auth.ErrUnauthorized
	}

	return next(ctx)
}

func HasPermission(ctx context.Context, obj interface{}, next graphql.Resolver, permission models.Permission) (res interface{}, err error) {
	user := auth.UserFromContext(ctx)
	if userHasAdminRole(user) {
		return next(ctx)
	}
	for _, rPermission := range user.Role.Permissions {
		if rPermission.Name == permission {
			return next(ctx)
		}
	}
	return nil, errors.New("user does not have permission.")
}
