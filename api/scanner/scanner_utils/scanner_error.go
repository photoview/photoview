package scanner_utils

import (
	"context"
	"fmt"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/graphql/notification"
	"github.com/photoview/photoview/api/log"
	"github.com/photoview/photoview/api/scanner/scanner_task"
	"github.com/photoview/photoview/api/utils"
)

func ScannerError(ctx context.Context, format string, args ...any) {
	message := fmt.Sprintf(format, args...)

	log.Error(ctx, message)
	note := &models.Notification{
		Key:      utils.GenerateToken(),
		Type:     models.NotificationTypeMessage,
		Header:   "Scanner error",
		Content:  message,
		Negative: true,
	}
	if taskCtx, ok := ctx.(scanner_task.TaskContext); ok && taskCtx.Context != nil {
		if album := taskCtx.GetAlbum(); album != nil {
			notification.BroadcastAlbumNotification(album.ID, note)
			return
		}
	}
	notification.BroadcastNotification(note)
}
