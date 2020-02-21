package resolvers

import (
	"context"

	"github.com/viktorstrate/photoview/api/graphql/auth"
	"github.com/viktorstrate/photoview/api/graphql/notification"
)

func (r *subscriptionResolver) Notification(ctx context.Context) (notification.NotificationChannel, error) {

	user := auth.UserFromContext(ctx)
	if user == nil {
		return nil, auth.ErrUnauthorized
	}

	notificationChannel := make(notification.NotificationChannel, 1)

	listenerID := notification.RegisterListener(user, notificationChannel)

	go func() {
		<-ctx.Done()
		notification.DeregisterListener(listenerID)
	}()

	return notificationChannel, nil
}
