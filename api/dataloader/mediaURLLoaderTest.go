package dataloader

import (
	"testing"

	"github.com/photoview/photoview/api/graphql/models"
	"go.uber.org/mock/gomock"
)

func TestNewThumbnailMediaURLLoader(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockMediaRepository(ctrl)
	mediaIDs := []int{1, 2}
	expectedURLs := []*models.MediaURL{
		{MediaID: 1, Purpose: models.PhotoThumbnail},
	}

	mockRepo.EXPECT().FindMediaURLs(gomock.Any(), mediaIDs).Return(expectedURLs, nil)

	loader := NewThumbnailMediaURLLoader(mockRepo)
	results, errs := loader.fetch(mediaIDs)

	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	if len(results) != len(mediaIDs) {
		t.Fatalf("expected %d results, got %d", len(mediaIDs), len(results))
	}

	for i, result := range results {
		if result != nil && result.MediaID != mediaIDs[i] {
			t.Errorf("expected MediaID %d, got %d", mediaIDs[i], result.MediaID)
		}
	}
}
