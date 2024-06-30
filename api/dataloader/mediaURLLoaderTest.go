package dataloader

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner/media_type"
	"gorm.io/gorm"
)

// MockMediaRepository is a mock of MediaRepository interface
type MockMediaRepository struct {
	mockCtrl *gomock.Controller
	mock     *dataloader.MockMediaRepository
}

func NewMockMediaRepository(ctrl *gomock.Controller) *MockMediaRepository {
	return &MockMediaRepository{
		mockCtrl: ctrl,
		mock:     dataloader.NewMockMediaRepository(ctrl),
	}
}

func (m *MockMediaRepository) FindMediaURLs(filter func(*gorm.DB) *gorm.DB, mediaIDs []int) ([]*models.MediaURL, error) {
	return m.mock.FindMediaURLs(filter, mediaIDs)
}

func TestNewThumbnailMediaURLLoader(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockMediaRepository(ctrl)
	mediaIDs := []int{1, 2}
	expectedURLs := []*models.MediaURL{
		{MediaID: 1, URL: "http://example.com/1", Purpose: models.PhotoThumbnail},
	}

	mockRepo.mock.EXPECT().FindMediaURLs(gomock.Any(), mediaIDs).Return(expectedURLs, nil)

	loader := dataloader.NewThumbnailMediaURLLoader(mockRepo)
	results, errs := loader.Fetch(mediaIDs)

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
