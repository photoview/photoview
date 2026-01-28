package auth

import (
	"context"
	"errors"
	"net/http"
	"regexp"

	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/photoview/photoview/api/dataloader"
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/log"
	"gorm.io/gorm"
)

var ErrUnauthorized = errors.New("unauthorized")
var bearerRegex = regexp.MustCompile("^(?i)Bearer ([a-zA-Z0-9]{24})$")

const INVALID_AUTH_TOKEN = "invalid authorization token"
const INTERNAL_SERVER_ERROR = "internal server error"

// A private key for context that only this package can access. This is important
// to prevent collisions between different context uses
var userCtxKey = &contextKey{"user"}

type contextKey struct {
	name string
}

// Middleware decodes the share session cookie and packs the session into context
func Middleware(db *gorm.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			if tokenCookie, err := r.Cookie("auth-token"); err == nil {
				loaders := dataloader.For(r.Context())
				if loaders == nil {
					log.Error(r.Context(), "Dataloader not available in HTTP context")
					http.Error(w, INTERNAL_SERVER_ERROR, http.StatusInternalServerError)
					return
				}

				user, err := loaders.UserFromAccessToken.Load(tokenCookie.Value)
				// Check for dataloader errors (database failures, etc.)
				if err != nil {
					log.Error(r.Context(), "Error loading user from token", "error", err)
					http.Error(w, INVALID_AUTH_TOKEN, http.StatusUnauthorized)
					return
				}

				// If user is nil, the token doesn't exist or is invalid
				if user == nil {
					log.Error(r.Context(), "Token not found in database")
					http.Error(w, INVALID_AUTH_TOKEN, http.StatusUnauthorized)
					return
				}

				// put it in context
				ctx := AddUserToContext(r.Context(), user)

				// and call the next with our new context
				r = r.WithContext(ctx)
			} else {
				if r.URL.Path != "/api/healthz" { // skip logging for health
					log.Info(r.Context(), "Did not find auth-token cookie")
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

func AddUserToContext(ctx context.Context, user *models.User) context.Context {
	return context.WithValue(ctx, userCtxKey, user)
}

func TokenFromBearer(bearer *string) (*string, error) {
	matches := bearerRegex.FindStringSubmatch(*bearer)
	if len(matches) != 2 {
		return nil, errors.New("invalid bearer format")
	}

	token := matches[1]
	return &token, nil
}

// UserFromContext finds the user from the context. REQUIRES Middleware to have run.
func UserFromContext(ctx context.Context) *models.User {
	raw, _ := ctx.Value(userCtxKey).(*models.User)
	return raw
}

func AuthWebsocketInit() func(context.Context, transport.InitPayload) (context.Context, *transport.InitPayload, error) {
	return func(ctx context.Context, initPayload transport.InitPayload) (context.Context, *transport.InitPayload, error) {

		bearer, exists := initPayload["Authorization"].(string)
		if !exists {
			return ctx, nil, nil
		}

		token, err := TokenFromBearer(&bearer)
		if err != nil {
			log.Error(ctx, "Invalid bearer format (websocket)", "error", err)
			return nil, nil, err
		}

		loaders := dataloader.For(ctx)
		if loaders == nil {
			log.Error(ctx, "Dataloader not available in websocket context")
			return nil, nil, errors.New(INTERNAL_SERVER_ERROR)
		}

		user, err := loaders.UserFromAccessToken.Load(*token)
		if err != nil {
			log.Error(ctx, "Error loading user from token (websocket)", "error", err)
			return nil, nil, errors.New(INVALID_AUTH_TOKEN)
		}

		// Check if token exists in database
		if user == nil {
			log.Error(ctx, "Token not found in database (websocket)")
			return nil, nil, errors.New(INVALID_AUTH_TOKEN)
		}

		// put it in context
		userCtx := context.WithValue(ctx, userCtxKey, user)

		// and return it so the resolvers can see it
		// Return nil for the InitPayload acknowledgment (no custom ack payload needed)
		return userCtx, nil, nil
	}
}
