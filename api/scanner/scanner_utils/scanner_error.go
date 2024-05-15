package scanner_utils

import (
	"fmt"
	"log"

	"github.com/kkovaletp/photoview/api/graphql/models"
	"github.com/kkovaletp/photoview/api/graphql/notification"
	"github.com/kkovaletp/photoview/api/utils"
)

func ScannerError(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)

	log.Printf("ERROR: %s", message)
	notification.BroadcastNotification(&models.Notification{
		Key:      utils.GenerateToken(),
		Type:     models.NotificationTypeMessage,
		Header:   "Scanner error",
		Content:  message,
		Negative: true,
	})
}
