package scanner_tasks

import (
	"fmt"
	"time"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/graphql/notification"
	"github.com/photoview/photoview/api/scanner/media_encoding"
	"github.com/photoview/photoview/api/scanner/scanner_task"
	"github.com/photoview/photoview/api/utils"
)

type NotificationTask struct {
	scanner_task.ScannerTaskBase
	throttle utils.Throttle
	albumKey string
}

func NewNotificationTask() NotificationTask {
	notifyThrottle := utils.NewThrottle(500 * time.Millisecond)
	notifyThrottle.Trigger(nil)

	return NotificationTask{
		albumKey: utils.GenerateToken(),
		throttle: notifyThrottle,
	}
}

func (t NotificationTask) AfterMediaFound(ctx scanner_task.TaskContext, media *models.Media, newMedia bool) error {
	if newMedia {
		t.throttle.Trigger(func() {
			notification.BroadcastNotification(&models.Notification{
				Key:     t.albumKey,
				Type:    models.NotificationTypeMessage,
				Header:  fmt.Sprintf("Found new media in album '%s'", ctx.GetAlbum().Title),
				Content: fmt.Sprintf("Found %s", media.Path),
			})
		})
	}

	return nil
}

func (t NotificationTask) AfterProcessMedia(ctx scanner_task.TaskContext, mediaData *media_encoding.EncodeMediaData, updatedURLs []*models.MediaURL, mediaIndex int, mediaTotal int) error {
	if len(updatedURLs) > 0 {
		progress := float64(mediaIndex) / float64(mediaTotal) * 100.0
		notification.BroadcastNotification(&models.Notification{
			Key:      t.albumKey,
			Type:     models.NotificationTypeProgress,
			Header:   fmt.Sprintf("Processing media for album '%s'", ctx.GetAlbum().Title),
			Content:  fmt.Sprintf("Processed media at %s", mediaData.Media.Path),
			Progress: &progress,
		})
	}

	return nil
}

func (t NotificationTask) AfterScanAlbum(ctx scanner_task.TaskContext, changedMedia []*models.Media, albumMedia []*models.Media) error {
	if len(changedMedia) > 0 {
		timeoutDelay := 2000
		notification.BroadcastNotification(&models.Notification{
			Key:      t.albumKey,
			Type:     models.NotificationTypeMessage,
			Positive: true,
			Header:   fmt.Sprintf("Done processing media for album '%s'", ctx.GetAlbum().Title),
			Content:  "All media have been processed",
			Timeout:  &timeoutDelay,
		})
	}

	return nil
}
