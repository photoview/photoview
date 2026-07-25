// Package helpers contains the helper functions for the Photoview API.
package helpers

import (
	"context"
	"fmt"

	"github.com/photoview/photoview/api/database"
	"github.com/photoview/photoview/api/graphql/models"
)

// SaveShareToken saves the share token to the database.
func SaveShareToken(ctx context.Context, token *models.ShareToken) error {
	// Save the share token to the database.
	return database.SaveShareTokenToDB(ctx, token)
}

// GetShareTokenFromDB returns the share token with the given ID from the database.
func GetShareTokenFromDB(ctx context.Context, id string) (*models.ShareToken, error) {
	// Get the share token from the database.
	return database.GetShareTokenFromDB(ctx, id)
}

// DeleteShareTokenFromDB deletes the share token with the given ID from the database.
func DeleteShareTokenFromDB(ctx context.Context, id string) error {
	// Delete the share token from the database.
	return database.DeleteShareTokenFromDB(ctx, id)
}