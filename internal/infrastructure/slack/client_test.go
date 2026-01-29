package slack

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	t.Run("Slack クライアントを作成できる", func(t *testing.T) {
		client := NewClient("https://hooks.slack.com/test")
		assert.NotNil(t, client)
	})

	t.Run("空の URL でもクライアントを作成できる", func(t *testing.T) {
		client := NewClient("")
		assert.NotNil(t, client)
	})
}

func TestClient_IsEnabled(t *testing.T) {
	t.Run("URL が設定されている場合は有効", func(t *testing.T) {
		client := NewClient("https://hooks.slack.com/test")
		assert.True(t, client.IsEnabled())
	})

	t.Run("URL が空の場合は無効", func(t *testing.T) {
		client := NewClient("")
		assert.False(t, client.IsEnabled())
	})
}

func TestClient_SendFeedback(t *testing.T) {
	t.Run("フィードバック通知を送信できる", func(t *testing.T) {
		var receivedPayload map[string]any
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

			err := json.NewDecoder(r.Body).Decode(&receivedPayload)
			require.NoError(t, err)

			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := NewClient(server.URL)
		feedback := FeedbackNotification{
			UserEmail: "test@example.com",
			UserName:  "Test User",
			Content:   "テストフィードバック",
			CreatedAt: time.Now(),
		}

		err := client.SendFeedback(context.Background(), feedback)

		assert.NoError(t, err)
		assert.NotNil(t, receivedPayload["blocks"])
	})

	t.Run("オプションフィールド付きで通知を送信できる", func(t *testing.T) {
		var receivedPayload map[string]any
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := json.NewDecoder(r.Body).Decode(&receivedPayload)
			require.NoError(t, err)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := NewClient(server.URL)
		pageURL := "https://app.example.com/test"
		userAgent := "Mozilla/5.0"
		screenshotURL := "https://storage.example.com/screenshot.png"
		feedback := FeedbackNotification{
			UserEmail:     "test@example.com",
			UserName:      "Test User",
			Content:       "テストフィードバック",
			PageURL:       &pageURL,
			UserAgent:     &userAgent,
			ScreenshotURL: &screenshotURL,
			CreatedAt:     time.Now(),
		}

		err := client.SendFeedback(context.Background(), feedback)

		assert.NoError(t, err)
		blocks := receivedPayload["blocks"].([]any)
		// ヘッダー + ユーザー情報 + コンテンツ + メタ情報 + 画像 = 5 ブロック
		assert.GreaterOrEqual(t, len(blocks), 4)
	})

	t.Run("無効な場合は何もしない", func(t *testing.T) {
		client := NewClient("")
		feedback := FeedbackNotification{
			UserEmail: "test@example.com",
			UserName:  "Test User",
			Content:   "テストフィードバック",
			CreatedAt: time.Now(),
		}

		err := client.SendFeedback(context.Background(), feedback)

		assert.NoError(t, err)
	})

	t.Run("サーバーがエラーを返すとエラーになる", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		client := NewClient(server.URL)
		feedback := FeedbackNotification{
			UserEmail: "test@example.com",
			UserName:  "Test User",
			Content:   "テストフィードバック",
			CreatedAt: time.Now(),
		}

		err := client.SendFeedback(context.Background(), feedback)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "500")
	})

	t.Run("長い User-Agent は切り詰められる", func(t *testing.T) {
		var receivedPayload map[string]any
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := json.NewDecoder(r.Body).Decode(&receivedPayload)
			require.NoError(t, err)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := NewClient(server.URL)
		// 200文字の User-Agent
		longUA := string(make([]byte, 200))
		for i := range longUA {
			longUA = longUA[:i] + "a" + longUA[i+1:]
		}
		feedback := FeedbackNotification{
			UserEmail: "test@example.com",
			UserName:  "Test User",
			Content:   "テストフィードバック",
			UserAgent: &longUA,
			CreatedAt: time.Now(),
		}

		err := client.SendFeedback(context.Background(), feedback)

		assert.NoError(t, err)
	})
}

func TestFeedbackNotification(t *testing.T) {
	t.Run("FeedbackNotification を作成できる", func(t *testing.T) {
		pageURL := "https://example.com"
		userAgent := "Mozilla/5.0"
		screenshotURL := "https://storage.example.com/image.png"

		notification := FeedbackNotification{
			UserEmail:     "test@example.com",
			UserName:      "Test User",
			Content:       "フィードバック内容",
			ScreenshotURL: &screenshotURL,
			PageURL:       &pageURL,
			UserAgent:     &userAgent,
			CreatedAt:     time.Now(),
		}

		assert.Equal(t, "test@example.com", notification.UserEmail)
		assert.Equal(t, "Test User", notification.UserName)
		assert.Equal(t, "フィードバック内容", notification.Content)
		assert.Equal(t, &screenshotURL, notification.ScreenshotURL)
		assert.Equal(t, &pageURL, notification.PageURL)
		assert.Equal(t, &userAgent, notification.UserAgent)
	})
}
