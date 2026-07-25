// Package database contains the database operations for the Photoview API.
package database

import (
	"context"
	"fmt"

	"github.com/photoview/photoview/api/graphql/models"
)

// CreateShareToken creates a new share token.
func CreateShareToken(ctx context.Context, name string, expiresAt int64, permissions []int) (*models.ShareToken, error) {
	// Create a new share token.
	token := &models.ShareToken{
		Name:      name,
		ExpiresAt: expiresAt,
		Permissions: permissions,
	}

	// Save the share token to the database.
	err := SaveShareToken(ctx, token)
	if err != nil {
		return nil, err
	}

	// Return the created share token.
	return token, nil
}

// GetShareToken returns the share token with the given ID.
func GetShareToken(ctx context.Context, id string) (*models.ShareToken, error) {
	// Get the share token from the database.
	token, err := GetShareTokenFromDB(ctx, id)
	if err != nil {
		return nil, err
	}

	// Return the share token.
	return token, nil
}

// UpdateShareToken updates the share token with the given ID.
func UpdateShareToken(ctx context.Context, id string, name string, expiresAt int64, permissions []int) (*models.ShareToken, error) {
	// Get the share token from the database.
	token, err := GetShareTokenFromDB(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update the share token.
	token.Name = name
	token.ExpiresAt = expiresAt
	token.Permissions = permissions

	// Save the updated share token to the database.
	err = SaveShareToken(ctx, token)
	if err != nil {
		return nil, err
	}

	// Return the updated share token.
	return token, nil
}

// DeleteShareToken deletes the share token with the given ID.
func DeleteShareToken(ctx context.Context, id string) error {
	// Delete the share token from the database.
	return DeleteShareTokenFromDB(ctx, id)
}