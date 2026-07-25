// Package models contains the GraphQL models for the Photoview API.
package models

import (
	"context"
	"fmt"
	"strings"

	"github.com/photoview/photoview/api/database"
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/graphql/resolvers"
)

// ShareToken represents a share token.
type ShareToken struct {
	ID        string `json:"id"`
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expiresAt"`
	// The name of the share token.
	Name string `json:"name"`
}

// ShareTokenInput represents the input for creating a share token.
type ShareTokenInput struct {
	Name        string `json:"name"`
	ExpiresAt   int64  `json:"expiresAt"`
	Permissions []int  `json:"permissions"`
}

// CreateShareToken creates a new share token.
func (m *models) CreateShareToken(ctx context.Context, input ShareTokenInput) (*ShareToken, error) {
	// Create a new share token.
	token, err := database.CreateShareToken(ctx, input.Name, input.ExpiresAt, input.Permissions)
	if err != nil {
		return nil, err
	}

	// Return the created share token.
	return &ShareToken{
		ID:        token.ID,
		Token:     token.Token,
		ExpiresAt: token.ExpiresAt,
		Name:      token.Name,
	}, nil
}