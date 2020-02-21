package notification

import (
	"errors"
	"github.com/viktorstrate/photoview/api/graphql/models"
	"log"
	"sync"
)

type NotificationChannel = <-chan *models.Notification

type NotificationListener struct {
	listenerID int
	user       models.User
	channel    NotificationChannel
}

func NewListener(user models.User, channel NotificationChannel) *NotificationListener {
	nextNotificationId++
	return &NotificationListener{
		listenerID: nextNotificationId,
		user:       user,
		channel:    channel,
	}
}

var notificationListeners []*NotificationListener = make([]*NotificationListener, 0)
var nextNotificationId = 0
var notificationLock = &sync.Mutex{}

func RegisterListener(user *models.User, channel NotificationChannel) int {
	log.Println("Registering notification listener")

	notificationLock.Lock()
	defer notificationLock.Unlock()

	notificationListeners = append(notificationListeners, NewListener(*user, channel))
	return nextNotificationId
}

func DeregisterListener(listenerID int) error {

	notificationLock.Lock()
	defer notificationLock.Unlock()

	for i, listener := range notificationListeners {
		if listener.listenerID == listenerID {

			if len(notificationListeners) > 1 {
				lastIndex := len(notificationListeners) - 1
				lastListener := notificationListeners[lastIndex]
				notificationListeners[i] = lastListener
				notificationListeners[lastIndex] = nil
				notificationListeners = notificationListeners[:lastIndex]
			} else {
				notificationListeners = make([]*NotificationListener, 0)
			}

			return nil
		}
	}

	return errors.New("ListenerID not found, while trying to deregister it")
}
