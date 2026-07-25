// Package resolvers contains the GraphQL resolvers for the Photoview API.
package resolvers

import (
	"context"
	"fmt"

	"github.com/photoview/photoview/api/database"
	"github.com/photoview/photoview/api/graphql/models"
)

// ShareTokenResolver represents a share token resolver.
type ShareTokenResolver struct {
	*models.ModelResolver
}

// GetShareToken returns the share token with the given ID.
func (r *ShareTokenResolver) GetShareToken(ctx context.Context, id string) (*models.ShareToken, error) {
	// Get the share token from the database.
	token, err := database.GetShareToken(ctx, id)
	if err != nil {
		return nil, err
	}

	// Return the share token.
	return &models.ShareToken{
		ID:        token.ID,
		Token:     token.Token,
		ExpiresAt: token.ExpiresAt,
		Name:      token.Name,
	}, nil
}

// UpdateShareToken updates the share token with the given ID.
func (r *ShareTokenResolver) UpdateShareToken(ctx context.Context, id string, input models.ShareTokenInput) (*models.ShareToken, error) {
	// Update the share token in the database.
	token, err := database.UpdateShareToken(ctx, id, input.Name, input.ExpiresAt, input.Permissions)
	if err != nil {
		return nil, err
	}

	// Return the updated share token.
	return &models.ShareToken{
		ID:        token.ID,
		Token:     token.Token,
		ExpiresAt: token.ExpiresAt,
		Name:      token.Name,
	}, nil
}

// DeleteShareToken deletes the share token with the given ID.
func (r *ShareTokenResolver) DeleteShareToken(ctx context.Context, id string) error {
	// Delete the share token from the database.
	return database.DeleteShareToken(ctx, id)
}