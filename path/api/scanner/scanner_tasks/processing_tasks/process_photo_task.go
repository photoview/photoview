// Package processing_tasks contains the processing tasks for the Photoview API.
package processing_tasks

import (
	"context"
	"fmt"

	"github.com/photoview/photoview/api/database"
	"github.com/photoview/photoview/api/graphql/models"
)

// ProcessPhoto processes a photo.
func ProcessPhoto(ctx context.Context, photo *models.Photo) error {
	// Process the photo.
	return database.ProcessPhoto(ctx, photo)
}