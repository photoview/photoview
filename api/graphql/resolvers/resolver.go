package resolvers

import (
	"context"

	"github.com/photoview/photoview/api/scanner/queue"
	"gorm.io/gorm"
)

//go:generate go tool github.com/99designs/gqlgen

type Resolver struct {
	database *gorm.DB
	queue    *queue.Queue
}

func NewRootResolver(db *gorm.DB, queue *queue.Queue) Resolver {
	return Resolver{
		database: db,
		queue:    queue,
	}
}

// DB returns a database instance that is tied to the given context
func (r *Resolver) DB(ctx context.Context) *gorm.DB {
	return r.database.WithContext(ctx)
}
