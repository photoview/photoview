// Package routes contains the routes for the Photoview API.
package routes

import (
	"context"
	"fmt"

	"github.com/photoview/photoview/api/database"
	"github.com/photoview/photoview/api/graphql/models"
)

// CreateShareToken creates a new share token.
func CreateShareToken(ctx context.Context, name string, expiresAt int64, permissions []int) (*models.ShareToken, error) {
	// Create a new share token.
	token, err := database.CreateShareToken(ctx, name, expiresAt, permissions)
	if err != nil {
		return nil, err
	}

	// Return the created share token.
	return token, nil
}

// GetShareToken returns the share token with the given ID.
func GetShareToken(ctx context.Context, id string) (*models.ShareToken, error) {
	// Get the share token from the database.
	token, err := database.GetShareToken(ctx, id)
	if err != nil {
		return nil, err
	}

	// Return the share token.
	return token, nil
}

// UpdateShareToken updates the share token with the given ID.
func UpdateShareToken(ctx context.Context, id string, name string, expiresAt int64, permissions []int) (*models.ShareToken, error) {
	// Update the share token in the database.
	token, err := database.UpdateShareToken(ctx, id, name, expiresAt, permissions)
	if err != nil {
		return nil, err
	}

	// Return the updated share token.
	return token, nil
}

// DeleteShareToken deletes the share token with the given ID.
func DeleteShareToken(ctx context.Context, id string) error {
	// Delete the share token from the database.
	return database.DeleteShareToken(ctx, id)
}