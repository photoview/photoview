package resolvers

import (
	"context"

	"github.com/kkovaletp/photoview/api/graphql/auth"
	"github.com/kkovaletp/photoview/api/graphql/models"
	"github.com/kkovaletp/photoview/api/graphql/notification"
)

func (r *subscriptionResolver) Notification(ctx context.Context) (<-chan *models.Notification, error) {

	user := auth.UserFromContext(ctx)
	if user == nil {
		return nil, auth.ErrUnauthorized
	}

	notificationChannel := make(chan *models.Notification, 1)

	listenerID := notification.RegisterListener(user, notificationChannel)

	go func() {
		<-ctx.Done()
		notification.DeregisterListener(listenerID)
	}()

	return notificationChannel, nil
}
