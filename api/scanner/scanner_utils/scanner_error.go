package scanner_utils

import (
	"context"
	"fmt"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/graphql/notification"
	"github.com/photoview/photoview/api/log"
	"github.com/photoview/photoview/api/utils"
)

func ScannerError(ctx context.Context, format string, args ...any) {
	message := fmt.Sprintf(format, args...)

	log.Error(ctx, message)
	notification.BroadcastNotification(&models.Notification{
		Key:      utils.GenerateToken(),
		Type:     models.NotificationTypeMessage,
		Header:   "Scanner error",
		Content:  message,
		Negative: true,
	})
}
