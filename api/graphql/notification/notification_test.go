package notification

import (
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/stretchr/testify/require"
)

func resetNotificationListeners(t *testing.T) {
	t.Helper()
	notificationLock.Lock()
	defer notificationLock.Unlock()
	for _, listener := range notificationListeners {
		listener.retire()
	}
	notificationListeners = nil
}

func TestDeregisterClosesDoneOnce(t *testing.T) {
	resetNotificationListeners(t)

	user := models.User{Username: "alice"}
	id := RegisterListener(&user, make(chan *models.Notification, 1))

	notificationLock.Lock()
	listener := notificationListeners[0]
	notificationLock.Unlock()

	require.NoError(t, DeregisterListener(id))
	listener.retire()

	select {
	case <-listener.done:
	default:
		t.Fatal("done channel was not closed")
	}
	require.EqualError(t, DeregisterListener(id), "ListenerID not found, while trying to deregister it")
}

func TestBroadcastWaitsForActiveListener(t *testing.T) {
	resetNotificationListeners(t)

	user := models.User{Username: "alice"}
	ch := make(chan *models.Notification, 1)
	ch <- &models.Notification{Key: "old"}
	id := RegisterListener(&user, ch)
	t.Cleanup(func() { _ = DeregisterListener(id) })

	finished := make(chan struct{})
	go func() {
		BroadcastNotification(&models.Notification{Key: "new"})
		close(finished)
	}()

	select {
	case <-finished:
		t.Fatal("broadcast returned while the active listener's channel was full")
	case <-time.After(50 * time.Millisecond):
	}

	select {
	case got := <-ch:
		require.Equal(t, "old", got.Key)
	case <-time.After(2 * time.Second):
		t.Fatal("timed out reading the buffered notification")
	}

	select {
	case <-finished:
	case <-time.After(2 * time.Second):
		t.Fatal("broadcast did not deliver to the active listener")
	}

	select {
	case got := <-ch:
		require.Equal(t, "new", got.Key)
	case <-time.After(2 * time.Second):
		t.Fatal("active listener did not receive the notification")
	}
}

func TestBroadcastExitsWhenDeregisteredListenerChannelIsFull(t *testing.T) {
	resetNotificationListeners(t)

	user := models.User{Username: "alice"}

	// One-slot channel, already full: a plain send would block forever.
	blockedCh := make(chan *models.Notification, 1)
	blockedCh <- &models.Notification{Key: "occupied"}
	laterCh := make(chan *models.Notification, 1)

	entered := make(chan struct{})
	release := make(chan struct{})
	var releaseOnce sync.Once
	releaseAllow := func() { releaseOnce.Do(func() { close(release) }) }

	blockedID := RegisterListenerWithAlbumAccess(&user, blockedCh, func(int) bool {
		close(entered)
		<-release
		return true
	})
	laterID := RegisterListener(&user, laterCh)
	t.Cleanup(func() {
		releaseAllow()
		_ = DeregisterListener(blockedID)
		_ = DeregisterListener(laterID)
	})

	finished := make(chan struct{})
	go func() {
		BroadcastAlbumNotification(9, &models.Notification{Key: "live"})
		close(finished)
	}()

	select {
	case <-entered:
	case <-time.After(2 * time.Second):
		t.Fatal("broadcast did not reach the listener")
	}

	// Remove the listener while broadcast is still on it, then let the send run.
	require.NoError(t, DeregisterListener(blockedID))
	releaseAllow()

	select {
	case <-finished:
	case <-time.After(2 * time.Second):
		t.Fatal("broadcast hung after the listener was deregistered")
	}

	select {
	case got := <-laterCh:
		require.Equal(t, "live", got.Key)
	default:
		t.Fatal("later listener received nothing")
	}

	select {
	case got := <-blockedCh:
		require.Equal(t, "occupied", got.Key)
	default:
		t.Fatal("retired listener lost its buffered notification")
	}
}

// TestDeregisterWhileBroadcastSendIsBlocked removes a listener only after
// broadcast is parked in its send select. The predecessor receive orders the
// test behind that listener; the stack shows the send is blocked. A timeout
// fails the test if deregister does not wake it.
func TestDeregisterWhileBroadcastSendIsBlocked(t *testing.T) {
	resetNotificationListeners(t)

	user := models.User{Username: "alice"}

	passedCh := make(chan *models.Notification, 1)
	blockedCh := make(chan *models.Notification, 1)
	blockedCh <- &models.Notification{Key: "occupied"}
	laterCh := make(chan *models.Notification, 1)

	passedID := RegisterListener(&user, passedCh)
	blockedID := RegisterListener(&user, blockedCh)
	laterID := RegisterListener(&user, laterCh)
	t.Cleanup(func() {
		_ = DeregisterListener(passedID)
		_ = DeregisterListener(blockedID)
		_ = DeregisterListener(laterID)
	})

	finished := make(chan struct{})
	go func() {
		BroadcastNotification(&models.Notification{Key: "live"})
		close(finished)
	}()

	select {
	case got := <-passedCh:
		require.Equal(t, "live", got.Key)
	case <-time.After(2 * time.Second):
		t.Fatal("broadcast did not reach the listener before the blocked one")
	}

	waitUntilBroadcastSelectBlocked(t, 2*time.Second)

	select {
	case <-finished:
		t.Fatal("broadcast returned while the send was still blocked")
	default:
	}

	require.NoError(t, DeregisterListener(blockedID))

	select {
	case <-finished:
	case <-time.After(2 * time.Second):
		t.Fatal("broadcast stayed blocked after the listener was deregistered")
	}

	select {
	case got := <-laterCh:
		require.Equal(t, "live", got.Key)
	default:
		t.Fatal("later listener received nothing")
	}

	select {
	case got := <-blockedCh:
		require.Equal(t, "occupied", got.Key)
	default:
		t.Fatal("blocked listener lost its buffered notification")
	}
}

// waitUntilBroadcastSelectBlocked returns once a broadcast goroutine is parked
// in select. Yielding observes that park; the deadline only bounds failure.
func waitUntilBroadcastSelectBlocked(t *testing.T, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	buf := make([]byte, 1<<20)
	for {
		n := runtime.Stack(buf, true)
		if broadcastParkedInSelect(buf[:n]) {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("broadcast send did not block in select:\n%s", buf[:n])
		}
		runtime.Gosched()
	}
}

func broadcastParkedInSelect(stack []byte) bool {
	for _, g := range strings.Split(string(stack), "\n\n") {
		header, rest, ok := strings.Cut(g, "\n")
		if !ok || !strings.Contains(header, "[select]") {
			continue
		}
		if strings.Contains(rest, "notification.broadcast(") ||
			strings.Contains(rest, "notification.BroadcastNotification(") ||
			strings.Contains(rest, "notification.BroadcastAlbumNotification(") {
			return true
		}
	}
	return false
}
