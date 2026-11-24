package notification

import (
	"errors"
	"log"
	"sync"

	"github.com/photoview/photoview/api/graphql/models"
)

type NotificationChannel = chan<- *models.Notification

type NotificationListener struct {
	listenerID int
	user       models.User
	channel    NotificationChannel
	// done is closed when the listener is deregistered. A broadcast that already
	// copied the listener selects on it so a send to a full channel can exit.
	done     chan struct{}
	doneOnce sync.Once
	// allow reports whether this listener may receive a notification for an
	// album. Nil allows every notification. The check runs when the notification
	// is sent, so a websocket subscription can stop receiving an album after
	// access to it is revoked.
	allow func(albumID int) bool
}

func NewListener(user models.User, channel NotificationChannel) *NotificationListener {
	nextNotificationId++
	return &NotificationListener{
		listenerID: nextNotificationId,
		user:       user,
		channel:    channel,
		done:       make(chan struct{}),
	}
}

// retire closes done exactly once so a repeated deregister cannot panic.
func (l *NotificationListener) retire() {
	l.doneOnce.Do(func() {
		close(l.done)
	})
}

var notificationListeners []*NotificationListener = make([]*NotificationListener, 0)
var nextNotificationId = 0
var notificationLock = &sync.Mutex{}

func RegisterListener(user *models.User, channel NotificationChannel) int {
	return registerListener(user, channel, nil)
}

// RegisterListenerWithAlbumAccess registers a listener that re-checks album
// access for every album-scoped notification. Global notifications, which have
// no album, are still delivered.
func RegisterListenerWithAlbumAccess(user *models.User, channel NotificationChannel, allow func(albumID int) bool) int {
	return registerListener(user, channel, allow)
}

func registerListener(user *models.User, channel NotificationChannel, allow func(albumID int) bool) int {
	log.Println("Registering notification listener")

	notificationLock.Lock()
	defer notificationLock.Unlock()

	listener := NewListener(*user, channel)
	listener.allow = allow
	notificationListeners = append(notificationListeners, listener)
	return nextNotificationId
}

func DeregisterListener(listenerID int) error {

	notificationLock.Lock()
	defer notificationLock.Unlock()

	for i, listener := range notificationListeners {

		log.Println("Deregistering notification listener")

		if listener.listenerID == listenerID {
			listener.retire()

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

func BroadcastNotification(note *models.Notification) {
	broadcast(note, 0)
}

// BroadcastAlbumNotification delivers note only to listeners who may currently
// see albumID. Listeners registered without an access check receive it as well.
func BroadcastAlbumNotification(albumID int, note *models.Notification) {
	broadcast(note, albumID)
}

func broadcast(note *models.Notification, albumID int) {
	if note == nil {
		return
	}

	notificationLock.Lock()
	listeners := append([]*NotificationListener(nil), notificationListeners...)
	notificationLock.Unlock()

	for _, listener := range listeners {
		if albumID != 0 && listener.allow != nil && !listener.allow(albumID) {
			continue
		}
		select {
		case listener.channel <- note:
		case <-listener.done:
		}
	}
}
