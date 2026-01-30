package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/siropaca/anycast-backend/internal/pkg/logger"
)

// Client は Slack への通知を行うインターフェース
type Client interface {
	SendFeedback(ctx context.Context, feedback FeedbackNotification) error
	IsEnabled() bool
}

// FeedbackNotification はフィードバック通知の内容を表す
type FeedbackNotification struct {
	UserEmail     string
	UserName      string
	Content       string
	ScreenshotURL *string
	PageURL       *string
	UserAgent     *string
	CreatedAt     time.Time
}

type slackClient struct {
	webhookURL string
	httpClient *http.Client
}

// NewClient は Slack クライアントを生成する
// webhookURL が空の場合は通知が無効化される
func NewClient(webhookURL string) Client {
	return &slackClient{
		webhookURL: webhookURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// IsEnabled は Slack 通知が有効かどうかを返す
func (c *slackClient) IsEnabled() bool {
	return c.webhookURL != ""
}

// SendFeedback はフィードバック通知を Slack に送信する
func (c *slackClient) SendFeedback(ctx context.Context, feedback FeedbackNotification) error {
	if !c.IsEnabled() {
		return nil
	}

	blocks := []map[string]any{
		{
			"type": "header",
			"text": map[string]string{
				"type": "plain_text",
				"text": "New Feedback Received",
			},
		},
		{
			"type": "section",
			"fields": []map[string]string{
				{"type": "mrkdwn", "text": fmt.Sprintf("*User:*\n%s (%s)", feedback.UserName, feedback.UserEmail)},
				{"type": "mrkdwn", "text": fmt.Sprintf("*Date:*\n%s", feedback.CreatedAt.Format(time.RFC3339))},
			},
		},
		{
			"type": "section",
			"text": map[string]string{
				"type": "mrkdwn",
				"text": fmt.Sprintf("*Content:*\n%s", feedback.Content),
			},
		},
	}

	// メタ情報を追加
	var metaFields []map[string]string
	if feedback.PageURL != nil && *feedback.PageURL != "" {
		metaFields = append(metaFields, map[string]string{
			"type": "mrkdwn",
			"text": fmt.Sprintf("*Page URL:*\n%s", *feedback.PageURL),
		})
	}
	if feedback.UserAgent != nil && *feedback.UserAgent != "" {
		// User-Agent は長いので切り詰め
		ua := *feedback.UserAgent
		if len(ua) > 100 {
			ua = ua[:100] + "..."
		}
		metaFields = append(metaFields, map[string]string{
			"type": "mrkdwn",
			"text": fmt.Sprintf("*User-Agent:*\n%s", ua),
		})
	}
	if len(metaFields) > 0 {
		blocks = append(blocks, map[string]any{
			"type":   "section",
			"fields": metaFields,
		})
	}

	// スクリーンショットがある場合
	if feedback.ScreenshotURL != nil && *feedback.ScreenshotURL != "" {
		blocks = append(blocks, map[string]any{
			"type":      "image",
			"image_url": *feedback.ScreenshotURL,
			"alt_text":  "Screenshot",
		})
	}

	payload := map[string]any{
		"blocks": blocks,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal slack payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.webhookURL, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create slack request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send slack notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.FromContext(ctx).Warn("slack notification failed", "status", resp.StatusCode)
		return fmt.Errorf("slack returned status %d", resp.StatusCode)
	}

	return nil
}
