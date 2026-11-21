package graphql_endpoint_test

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/photoview/photoview/api/dataloader"
	graphql_endpoint "github.com/photoview/photoview/api/graphql/endpoint"
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/graphql/notification"
	"github.com/photoview/photoview/api/test_utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

const notificationSubscription = "subscription { notification { key header content } }"

// An open websocket subscription must stop delivering album notifications after
// that album membership is removed. The connection stays open.
func TestWebsocketSubscriptionStopsDeliveringAfterAlbumRevocation(t *testing.T) {
	db := test_utils.DatabaseTest(t)
	password := "1234"
	alice, err := models.RegisterUser(db, "alice", &password, false)
	require.NoError(t, err)
	bob, err := models.RegisterUser(db, "bob", &password, false)
	require.NoError(t, err)
	admin, err := models.RegisterUser(db, "admin", &password, true)
	require.NoError(t, err)

	aliceAlbum := models.Album{Title: "alice", Path: "/photos/alice"}
	require.NoError(t, db.Save(&aliceAlbum).Error)
	require.NoError(t, db.Model(alice).Association("Albums").Append(&aliceAlbum))
	bobAlbum := models.Album{Title: "bob", Path: "/photos/bob"}
	require.NoError(t, db.Save(&bobAlbum).Error)
	require.NoError(t, db.Model(bob).Association("Albums").Append(&bobAlbum))

	t.Setenv("PHOTOVIEW_DEVELOPMENT_MODE", "1")
	server := httptest.NewServer(dataloader.Middleware(db)(graphql_endpoint.GraphqlEndpoint(db)))
	t.Cleanup(server.Close)
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	aliceConn := subscribeNotifications(t, wsURL, accessToken(t, db, alice))
	bobConn := subscribeNotifications(t, wsURL, accessToken(t, db, bob))
	adminConn := subscribeNotifications(t, wsURL, accessToken(t, db, admin))
	awaitSubscription(t, aliceConn, bobConn, adminConn)

	// Bob's album is delivered to Bob and to an administrator, not to Alice.
	notification.BroadcastAlbumNotification(bobAlbum.ID, &models.Notification{
		Key: "bob-1", Type: models.NotificationTypeMessage, Header: "bob", Content: "one",
	})
	assert.Equal(t, "bob-1", waitForNotificationKey(t, bobConn, 2*time.Second))
	assert.Equal(t, "bob-1", waitForNotificationKey(t, adminConn, 2*time.Second))
	assert.Empty(t, waitForNotificationKey(t, aliceConn, 400*time.Millisecond))

	// Alice receives her own album on this same connection.
	notification.BroadcastAlbumNotification(aliceAlbum.ID, &models.Notification{
		Key: "alice-1", Type: models.NotificationTypeMessage, Header: "alice", Content: "one",
	})
	assert.Equal(t, "alice-1", waitForNotificationKey(t, aliceConn, 2*time.Second))
	assert.Equal(t, "alice-1", waitForNotificationKey(t, adminConn, 2*time.Second))
	assert.Empty(t, waitForNotificationKey(t, bobConn, 400*time.Millisecond))

	require.NoError(t, db.Exec("DELETE FROM user_albums WHERE user_id = ?", alice.ID).Error)

	// The subscription was not restarted. It stops delivering the revoked album.
	notification.BroadcastAlbumNotification(aliceAlbum.ID, &models.Notification{
		Key: "alice-2", Type: models.NotificationTypeMessage, Header: "alice", Content: "two",
	})
	assert.Equal(t, "alice-2", waitForNotificationKey(t, adminConn, 2*time.Second))
	assert.Empty(t, waitForNotificationKey(t, aliceConn, 400*time.Millisecond))
	assert.Empty(t, waitForNotificationKey(t, bobConn, 400*time.Millisecond))

	assert.Equal(t, "alice", queryUsername(t, aliceConn))
}

func accessToken(t *testing.T, db *gorm.DB, user *models.User) string {
	t.Helper()
	token, err := user.GenerateAccessToken(db)
	require.NoError(t, err)
	return token.Value
}

type wsClient struct {
	conn *websocket.Conn
	msgs <-chan wsMessage
}

func subscribeNotifications(t *testing.T, url string, token string) *wsClient {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	conn, resp, err := websocket.Dial(ctx, url, &websocket.DialOptions{
		Subprotocols: []string{"graphql-ws"},
	})
	require.NoError(t, err)
	if resp != nil && resp.Body != nil {
		resp.Body.Close()
	}
	t.Cleanup(func() { _ = conn.Close(websocket.StatusNormalClosure, "done") })

	client := &wsClient{conn: conn}
	client.write(t, map[string]any{
		"type":    "connection_init",
		"payload": map[string]any{"Authorization": "Bearer " + token},
	})
	ack := client.readOne(t, 5*time.Second)
	require.NotNil(t, ack)
	require.Equal(t, "connection_ack", ack.Type, "handshake response: %s", ack.raw)

	client.write(t, map[string]any{
		"id":   "sub",
		"type": "start",
		"payload": map[string]string{
			"query": notificationSubscription,
		},
	})
	client.msgs = client.pump(t)
	return client
}

func (c *wsClient) write(t *testing.T, payload any) {
	t.Helper()
	body, err := json.Marshal(payload)
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	require.NoError(t, c.conn.Write(ctx, websocket.MessageText, body))
}

type wsMessage struct {
	ID      string          `json:"id"`
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
	raw     string
}

func (c *wsClient) readOne(t *testing.T, timeout time.Duration) *wsMessage {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	_, data, err := c.conn.Read(ctx)
	if err != nil {
		return nil
	}
	msg, err := decodeWS(data)
	require.NoError(t, err)
	return &msg
}

// pump reads the socket until it closes. Reads are not canceled on a timeout,
// so a negative assertion cannot drop a frame that arrives a moment later.
func (c *wsClient) pump(t *testing.T) <-chan wsMessage {
	t.Helper()
	msgs := make(chan wsMessage, 16)
	go func() {
		defer close(msgs)
		for {
			_, data, err := c.conn.Read(context.Background())
			if err != nil {
				return
			}
			msg, err := decodeWS(data)
			if err != nil {
				continue
			}
			msgs <- msg
		}
	}()
	return msgs
}

func decodeWS(data []byte) (wsMessage, error) {
	var msg wsMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return wsMessage{}, err
	}
	msg.raw = string(data)
	return msg, nil
}

func waitForNotificationKey(t *testing.T, client *wsClient, timeout time.Duration) string {
	t.Helper()
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	for {
		select {
		case <-timer.C:
			return ""
		case msg, ok := <-client.msgs:
			if !ok {
				return ""
			}
			switch msg.Type {
			case "ka", "connection_ack", "complete":
				continue
			case "error", "connection_error":
				t.Fatalf("websocket error: %s", msg.raw)
			case "data", "next":
				if key := notificationKey(&msg); key != "" {
					return key
				}
				t.Logf("ignoring data message without a notification key: %s", msg.raw)
			default:
				t.Fatalf("unexpected websocket message: %s", msg.raw)
			}
		}
	}
}

// awaitSubscription broadcasts until every connection has delivered one event.
// The server accepts the subscription asynchronously, so the first broadcast
// can happen before the listener exists.
func awaitSubscription(t *testing.T, clients ...*wsClient) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	ready := map[*wsClient]bool{}
	for time.Now().Before(deadline) && len(ready) < len(clients) {
		notification.BroadcastNotification(&models.Notification{
			Key: "ready", Type: models.NotificationTypeMessage, Header: "ready", Content: "ready",
		})
		for _, client := range clients {
			if ready[client] {
				continue
			}
			if waitForNotificationKey(t, client, 200*time.Millisecond) == "ready" {
				ready[client] = true
			}
		}
	}
	require.Len(t, ready, len(clients), "subscriptions did not start delivering")
	for _, client := range clients {
		for waitForNotificationKey(t, client, 50*time.Millisecond) == "ready" {
		}
	}
}

func notificationKey(msg *wsMessage) string {
	var payload struct {
		Data struct {
			Notification struct {
				Key string `json:"key"`
			} `json:"notification"`
		} `json:"data"`
	}
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return ""
	}
	return payload.Data.Notification.Key
}

func queryUsername(t *testing.T, client *wsClient) string {
	t.Helper()
	client.write(t, map[string]any{
		"id":   "user",
		"type": "start",
		"payload": map[string]string{
			"query": "query { myUser { username } }",
		},
	})

	timer := time.NewTimer(2 * time.Second)
	defer timer.Stop()
	for {
		select {
		case <-timer.C:
			t.Fatal("timed out waiting for myUser on the open websocket")
		case msg, ok := <-client.msgs:
			if !ok {
				t.Fatal("websocket closed while waiting for myUser")
			}
			if key := notificationKey(&msg); key != "" && key != "ready" {
				t.Fatalf("subscription delivered %s after album access was revoked", key)
			}
			if msg.Type == "error" || msg.Type == "connection_error" {
				t.Fatalf("websocket error: %s", msg.raw)
			}
			if msg.ID != "user" || (msg.Type != "data" && msg.Type != "next") {
				continue
			}
			var payload struct {
				Data struct {
					MyUser struct {
						Username string `json:"username"`
					} `json:"myUser"`
				} `json:"data"`
			}
			require.NoError(t, json.Unmarshal(msg.Payload, &payload), "message: %s", msg.raw)
			if payload.Data.MyUser.Username != "" {
				return payload.Data.MyUser.Username
			}
		}
	}
}
