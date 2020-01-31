package auth

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"regexp"

	"github.com/viktorstrate/photoview/api/graphql/models"
)

var ErrUnauthorized = errors.New("unauthorized")

// A private key for context that only this package can access. This is important
// to prevent collisions between different context uses
var userCtxKey = &contextKey{"user"}

type contextKey struct {
	name string
}

// Middleware decodes the share session cookie and packs the session into context
func Middleware(db *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			bearer := r.Header.Get("Authorization")
			if bearer == "" {
				next.ServeHTTP(w, r)
				return
			}

			regex, _ := regexp.Compile("^Bearer ([a-zA-Z0-9]{24})$")
			matches := regex.FindStringSubmatch(bearer)
			if len(matches) != 2 {
				http.Error(w, "Invalid authorization header format", http.StatusBadRequest)
				return
			}

			token := matches[1]

			user, err := models.VerifyTokenAndGetUser(db, token)
			if err != nil {
				log.Printf("Invalid token: %s\n", err)
				http.Error(w, "Invalid authorization token", http.StatusForbidden)
				return
			}

			// put it in context
			ctx := context.WithValue(r.Context(), userCtxKey, user)

			// and call the next with our new context
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

// UserFromContext finds the user from the context. REQUIRES Middleware to have run.
func UserFromContext(ctx context.Context) *models.User {
	raw, _ := ctx.Value(userCtxKey).(*models.User)
	return raw
}
