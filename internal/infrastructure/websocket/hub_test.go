package websocket

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHub(t *testing.T) {
	t.Run("Hub を作成できる", func(t *testing.T) {
		hub := NewHub()

		assert.NotNil(t, hub)
		assert.NotNil(t, hub.clients)
		assert.NotNil(t, hub.register)
		assert.NotNil(t, hub.unregister)
	})
}

func TestHub_RegisterAndUnregister(t *testing.T) {
	t.Run("クライアントを登録・解除できる", func(t *testing.T) {
		hub := NewHub()
		go hub.Run()

		// クライアントを直接作成（conn は nil）
		client := &Client{
			hub:            hub,
			conn:           nil,
			userID:         "user-123",
			send:           make(chan []byte, 256),
			subscribedJobs: make(map[string]bool),
		}

		// 登録
		hub.register <- client
		time.Sleep(10 * time.Millisecond)

		hub.mu.RLock()
		_, exists := hub.clients["user-123"]
		hub.mu.RUnlock()
		assert.True(t, exists)

		// 解除
		hub.unregister <- client
		time.Sleep(10 * time.Millisecond)

		hub.mu.RLock()
		_, exists = hub.clients["user-123"]
		hub.mu.RUnlock()
		assert.False(t, exists)
	})

	t.Run("同じユーザーの複数クライアントを登録できる", func(t *testing.T) {
		hub := NewHub()
		go hub.Run()

		client1 := &Client{
			hub:            hub,
			userID:         "user-123",
			send:           make(chan []byte, 256),
			subscribedJobs: make(map[string]bool),
		}
		client2 := &Client{
			hub:            hub,
			userID:         "user-123",
			send:           make(chan []byte, 256),
			subscribedJobs: make(map[string]bool),
		}

		hub.register <- client1
		hub.register <- client2
		time.Sleep(10 * time.Millisecond)

		hub.mu.RLock()
		clients := hub.clients["user-123"]
		count := len(clients)
		hub.mu.RUnlock()

		assert.Equal(t, 2, count)
	})
}

func TestHub_SendToUser(t *testing.T) {
	t.Run("ユーザーにメッセージを送信できる", func(t *testing.T) {
		hub := NewHub()
		go hub.Run()

		client := &Client{
			hub:            hub,
			userID:         "user-123",
			send:           make(chan []byte, 256),
			subscribedJobs: make(map[string]bool),
		}

		hub.register <- client
		time.Sleep(10 * time.Millisecond)

		msg := Message{Type: "test", Payload: "hello"}
		hub.SendToUser("user-123", msg)

		select {
		case data := <-client.send:
			var received Message
			err := json.Unmarshal(data, &received)
			require.NoError(t, err)
			assert.Equal(t, "test", received.Type)
		case <-time.After(100 * time.Millisecond):
			t.Fatal("メッセージを受信できなかった")
		}
	})

	t.Run("存在しないユーザーへの送信は無視される", func(t *testing.T) {
		hub := NewHub()
		go hub.Run()

		msg := Message{Type: "test", Payload: "hello"}
		// パニックしないことを確認
		hub.SendToUser("non-existent-user", msg)
	})
}

func TestHub_SendToJob(t *testing.T) {
	t.Run("ジョブを購読しているクライアントにメッセージを送信できる", func(t *testing.T) {
		hub := NewHub()
		go hub.Run()

		client := &Client{
			hub:            hub,
			userID:         "user-123",
			send:           make(chan []byte, 256),
			subscribedJobs: make(map[string]bool),
		}

		hub.register <- client
		time.Sleep(10 * time.Millisecond)

		// ジョブを購読
		client.mu.Lock()
		client.subscribedJobs["job-456"] = true
		client.mu.Unlock()

		msg := Message{Type: "progress", Payload: map[string]interface{}{"progress": 50}}
		hub.SendToJob("job-456", msg)

		select {
		case data := <-client.send:
			var received Message
			err := json.Unmarshal(data, &received)
			require.NoError(t, err)
			assert.Equal(t, "progress", received.Type)
		case <-time.After(100 * time.Millisecond):
			t.Fatal("メッセージを受信できなかった")
		}
	})

	t.Run("購読していないジョブのメッセージは受信しない", func(t *testing.T) {
		hub := NewHub()
		go hub.Run()

		client := &Client{
			hub:            hub,
			userID:         "user-123",
			send:           make(chan []byte, 256),
			subscribedJobs: make(map[string]bool),
		}

		hub.register <- client
		time.Sleep(10 * time.Millisecond)

		msg := Message{Type: "progress", Payload: map[string]interface{}{"progress": 50}}
		hub.SendToJob("job-456", msg)

		select {
		case <-client.send:
			t.Fatal("購読していないジョブのメッセージを受信した")
		case <-time.After(50 * time.Millisecond):
			// 期待通り受信しない
		}
	})
}

func TestClient_HandleMessage(t *testing.T) {
	t.Run("subscribe メッセージでジョブを購読できる", func(t *testing.T) {
		hub := NewHub()
		client := &Client{
			hub:            hub,
			userID:         "user-123",
			send:           make(chan []byte, 256),
			subscribedJobs: make(map[string]bool),
		}

		msg := `{"type":"subscribe","payload":{"jobId":"job-123"}}`
		client.handleMessage([]byte(msg))

		client.mu.Lock()
		subscribed := client.subscribedJobs["job-123"]
		client.mu.Unlock()

		assert.True(t, subscribed)
	})

	t.Run("unsubscribe メッセージでジョブの購読を解除できる", func(t *testing.T) {
		hub := NewHub()
		client := &Client{
			hub:            hub,
			userID:         "user-123",
			send:           make(chan []byte, 256),
			subscribedJobs: map[string]bool{"job-123": true},
		}

		msg := `{"type":"unsubscribe","payload":{"jobId":"job-123"}}`
		client.handleMessage([]byte(msg))

		client.mu.Lock()
		_, exists := client.subscribedJobs["job-123"]
		client.mu.Unlock()

		assert.False(t, exists)
	})

	t.Run("ping メッセージで pong を返す", func(t *testing.T) {
		hub := NewHub()
		client := &Client{
			hub:            hub,
			userID:         "user-123",
			send:           make(chan []byte, 256),
			subscribedJobs: make(map[string]bool),
		}

		msg := `{"type":"ping"}`
		client.handleMessage([]byte(msg))

		select {
		case data := <-client.send:
			var received Message
			err := json.Unmarshal(data, &received)
			require.NoError(t, err)
			assert.Equal(t, "pong", received.Type)
		case <-time.After(100 * time.Millisecond):
			t.Fatal("pong を受信できなかった")
		}
	})

	t.Run("無効な JSON は無視される", func(t *testing.T) {
		hub := NewHub()
		client := &Client{
			hub:            hub,
			userID:         "user-123",
			send:           make(chan []byte, 256),
			subscribedJobs: make(map[string]bool),
		}

		// パニックしないことを確認
		client.handleMessage([]byte("invalid json"))
	})
}

func TestMessage_JSON(t *testing.T) {
	t.Run("Message を JSON にシリアライズできる", func(t *testing.T) {
		msg := Message{
			Type:    "progress",
			Payload: map[string]interface{}{"jobId": "123", "progress": 50},
		}

		data, err := json.Marshal(msg)

		require.NoError(t, err)
		assert.Contains(t, string(data), `"type":"progress"`)
		assert.Contains(t, string(data), `"jobId":"123"`)
	})

	t.Run("JSON を Message にデシリアライズできる", func(t *testing.T) {
		data := `{"type":"completed","payload":{"jobId":"456"}}`

		var msg Message
		err := json.Unmarshal([]byte(data), &msg)

		require.NoError(t, err)
		assert.Equal(t, "completed", msg.Type)
	})
}
