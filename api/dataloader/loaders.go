package dataloader

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

const loadersKey = "dataloaders"

type Loaders struct {
	MediaThumbnail      *MediaURLLoader
	MediaHighres        *MediaURLLoader
	MediaVideoWeb       *MediaURLLoader
	UserFromAccessToken *UserLoader
	UserMediaFavorite   *UserFavoritesLoader
}

func Middleware(db *gorm.DB) mux.MiddlewareFunc {
	repo := &GormMediaRepository{db: db}

	return mux.MiddlewareFunc(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			ctx := context.WithValue(r.Context(), loadersKey, &Loaders{
				MediaThumbnail:      NewThumbnailMediaURLLoader(repo),
				MediaHighres:        NewHighresMediaURLLoader(repo),
				MediaVideoWeb:       NewVideoWebMediaURLLoader(repo),
				UserFromAccessToken: NewUserLoaderByToken(db),
				UserMediaFavorite:   NewUserFavoriteLoader(db),
			})

			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	})
}

func For(ctx context.Context) *Loaders {
	return ctx.Value(loadersKey).(*Loaders)
}
