// Package database_drivers contains the database drivers for the Photoview API.
package database_drivers

import (
	"context"
	"fmt"

	"github.com/photoview/photoview/api/database"
	"github.com/photoview/photoview/api/graphql/models"
)

// SaveShareTokenToDB saves the share token to the database.
func SaveShareTokenToDB(ctx context.Context, token *models.ShareToken) error {
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