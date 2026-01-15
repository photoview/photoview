package resolvers

import (
	"context"

	"github.com/spf13/afero"
	"gorm.io/gorm"
)

//go:generate go tool github.com/99designs/gqlgen

type Resolver struct {
	database   *gorm.DB
	filesystem afero.Fs
}

func NewRootResolver(db *gorm.DB, fs afero.Fs) Resolver {
	return Resolver{
		database:   db,
		filesystem: fs,
	}
}

// DB returns a database instance that is tied to the given context
func (r *Resolver) DB(ctx context.Context) *gorm.DB {
	return r.database.WithContext(ctx)
}

func (r *Resolver) FS(ctx context.Context) afero.Fs {
	return r.filesystem
}
