package resolvers

import (
	"context"

	"github.com/spf13/afero"
	"gorm.io/gorm"
)

//go:generate go tool github.com/99designs/gqlgen

type Resolver struct {
	database *gorm.DB
	fileFs   afero.Fs
	cacheFs  afero.Fs
}

func NewRootResolver(db *gorm.DB, fileFs afero.Fs, cacheFs afero.Fs) Resolver {
	return Resolver{
		database: db,
		fileFs:   fileFs,
		cacheFs:  cacheFs,
	}
}

// DB returns a database instance that is tied to the given context
func (r *Resolver) DB(ctx context.Context) *gorm.DB {
	return r.database.WithContext(ctx)
}

func (r *Resolver) FileFS(ctx context.Context) afero.Fs {
	return r.fileFs
}

func (r *Resolver) CacheFS(ctx context.Context) afero.Fs {
	return r.cacheFs
}
