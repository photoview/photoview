package resolvers

import (
	"context"

	"gorm.io/gorm"
)

//go:generate go tool github.com/99designs/gqlgen

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
