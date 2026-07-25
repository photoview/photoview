// Package routes contains the tests for the Photoview API.
package routes

import (
	"context"
	"fmt"

	"github.com/photoview/photoview/api/database"
	"github.com/photoview/photoview/api/graphql/models"
)

func TestCreateShareToken(t *testing.T) {
	// Create a new share token.
	token, err := CreateShareToken(context.Background(), "test", 1643723400, []int{1})
	if err != nil {
		t.Errorf("CreateShareToken returned an error: %v", err)
	}

	// Check the created share token.
	if token.ID == "" {
		t.Errorf("Share token ID is empty")
	}

	// Delete the share token.
	DeleteShareToken(context.Background(), token.ID)
}

func TestGetShareToken(t *testing.T) {
	// Create a new share token.
	token, err := CreateShareToken(context.Background(), "test", 1643723400, []int{1})
	if err != nil {
		t.Errorf("CreateShareToken returned an error: %v", err)
	}

	// Get the share token from the database.
	dbToken, err := GetShareToken(context.Background(), token.ID)
	if err != nil {
		t.Errorf("GetShareToken returned an error: %v", err)
	}

	// Check the retrieved share token.
	if dbToken.ID != token.ID {
		t.Errorf("Share token ID does not match")
	}

	// Delete the share token.
	DeleteShareToken(context.Background(), token.ID)
}

func TestUpdateShareToken(t *testing.T) {
	// Create a new share token.
	token, err := CreateShareToken(context.Background(), "test", 1643723400, []int{1})
	if err != nil {
		t.Errorf("CreateShareToken returned an error: %v", err)
	}

	// Update the share token in the database.
	token, err = UpdateShareToken(context.Background(), token.ID, "updated", 1643723400, []int{1})
	if err != nil {
		t.Errorf("UpdateShareToken returned an error: %v", err)
	}

	// Get the updated share token from the database.
	dbToken, err := GetShareToken(context.Background(), token.ID)
	if err != nil {
		t.Errorf("GetShareToken returned an error: %v", err)
	}

	// Check the updated share token.
	if dbToken.Name != "updated" {
		t.Errorf("Share token name does not match")
	}

	// Delete the share token.
	DeleteShareToken(context.Background(), token.ID)
}

func TestDeleteShareToken(t *testing.T) {
	// Create a new share token.
	token, err := CreateShareToken(context.Background(), "test", 1643723400, []int{1})
	if err != nil {
		t.Errorf("CreateShareToken returned an error: %v", err)
	}

	// Delete the share token from the database.
	err = DeleteShareToken(context.Background(), token.ID)
	if err != nil {
		t.Errorf("DeleteShareToken returned an error: %v", err)
	}

	// Get the share token from the database.
	_, err = GetShareToken(context.Background(), token.ID)
	if err != nil {
		t.Errorf("GetShareToken returned an error: %v", err)
	}
}