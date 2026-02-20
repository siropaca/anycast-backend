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
	SendContact(ctx context.Context, contact ContactNotification) error
	SendAlert(ctx context.Context, alert AlertNotification) error
	SendRegistration(ctx context.Context, registration RegistrationNotification) error
	IsFeedbackEnabled() bool
	IsContactEnabled() bool
	IsAlertEnabled() bool
	IsRegistrationEnabled() bool
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

// AlertNotification はジョブ失敗時のアラート通知の内容を表す
type AlertNotification struct {
	JobID        string
	JobType      string
	ErrorCode    string
	ErrorMessage string
	OccurredAt   time.Time
}

// RegistrationNotification は新規ユーザー登録通知の内容を表す
type RegistrationNotification struct {
	UserID      string
	DisplayName string
	Email       string
	Method      string
	CreatedAt   time.Time
}

// ContactNotification はお問い合わせ通知の内容を表す
type ContactNotification struct {
	Category      string
	CategoryLabel string
	Email         string
	Name          string
	Content       string
	UserAgent     *string
	UserID        *string
	CreatedAt     time.Time
}

type slackClient struct {
	feedbackWebhookURL     string
	contactWebhookURL      string
	alertWebhookURL        string
	registrationWebhookURL string
	httpClient             *http.Client
}

// NewClient は Slack クライアントを生成する
//
// 各 Webhook URL が空の場合、対応する通知が無効化される
//
// feedbackWebhookURL: フィードバック通知用の Slack Webhook URL
// contactWebhookURL: お問い合わせ通知用の Slack Webhook URL
// alertWebhookURL: アラート通知用の Slack Webhook URL
// registrationWebhookURL: 新規登録通知用の Slack Webhook URL
func NewClient(feedbackWebhookURL, contactWebhookURL, alertWebhookURL, registrationWebhookURL string) Client {
	return &slackClient{
		feedbackWebhookURL:     feedbackWebhookURL,
		contactWebhookURL:      contactWebhookURL,
		alertWebhookURL:        alertWebhookURL,
		registrationWebhookURL: registrationWebhookURL,
		httpClient:             &http.Client{Timeout: 10 * time.Second},
	}
}

// IsFeedbackEnabled は Slack フィードバック通知が有効かどうかを返す
func (c *slackClient) IsFeedbackEnabled() bool {
	return c.feedbackWebhookURL != ""
}

// IsContactEnabled は Slack お問い合わせ通知が有効かどうかを返す
func (c *slackClient) IsContactEnabled() bool {
	return c.contactWebhookURL != ""
}

// IsAlertEnabled は Slack アラート通知が有効かどうかを返す
func (c *slackClient) IsAlertEnabled() bool {
	return c.alertWebhookURL != ""
}

// IsRegistrationEnabled は Slack 新規登録通知が有効かどうかを返す
func (c *slackClient) IsRegistrationEnabled() bool {
	return c.registrationWebhookURL != ""
}

// SendFeedback はフィードバック通知を Slack に送信する
func (c *slackClient) SendFeedback(ctx context.Context, feedback FeedbackNotification) error {
	if !c.IsFeedbackEnabled() {
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

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.feedbackWebhookURL, bytes.NewBuffer(body))
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

// SendContact はお問い合わせ通知を Slack に送信する
func (c *slackClient) SendContact(ctx context.Context, contact ContactNotification) error {
	if !c.IsContactEnabled() {
		return nil
	}

	blocks := []map[string]any{
		{
			"type": "header",
			"text": map[string]string{
				"type": "plain_text",
				"text": "New Contact Received",
			},
		},
		{
			"type": "section",
			"fields": []map[string]string{
				{"type": "mrkdwn", "text": fmt.Sprintf("*Category:*\n%s", contact.CategoryLabel)},
				{"type": "mrkdwn", "text": fmt.Sprintf("*Date:*\n%s", contact.CreatedAt.Format(time.RFC3339))},
			},
		},
		{
			"type": "section",
			"fields": []map[string]string{
				{"type": "mrkdwn", "text": fmt.Sprintf("*Name:*\n%s", contact.Name)},
				{"type": "mrkdwn", "text": fmt.Sprintf("*Email:*\n%s", contact.Email)},
			},
		},
		{
			"type": "section",
			"text": map[string]string{
				"type": "mrkdwn",
				"text": fmt.Sprintf("*Content:*\n%s", contact.Content),
			},
		},
	}

	// メタ情報を追加
	var metaFields []map[string]string
	if contact.UserID != nil && *contact.UserID != "" {
		metaFields = append(metaFields, map[string]string{
			"type": "mrkdwn",
			"text": fmt.Sprintf("*User ID:*\n%s", *contact.UserID),
		})
	}
	if contact.UserAgent != nil && *contact.UserAgent != "" {
		ua := *contact.UserAgent
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

	payload := map[string]any{
		"blocks": blocks,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal slack payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.contactWebhookURL, bytes.NewBuffer(body))
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

// SendAlert はジョブ失敗時のアラート通知を Slack に送信する
func (c *slackClient) SendAlert(ctx context.Context, alert AlertNotification) error {
	if !c.IsAlertEnabled() {
		return nil
	}

	// エラーメッセージが長い場合は切り詰め
	errorMessage := alert.ErrorMessage
	if len(errorMessage) > 300 {
		errorMessage = errorMessage[:300] + "..."
	}

	blocks := []map[string]any{
		{
			"type": "header",
			"text": map[string]string{
				"type": "plain_text",
				"text": "🚨 Job Failed Alert",
			},
		},
		{
			"type": "section",
			"fields": []map[string]string{
				{"type": "mrkdwn", "text": fmt.Sprintf("*Job Type:*\n%s", alert.JobType)},
				{"type": "mrkdwn", "text": fmt.Sprintf("*Error Code:*\n%s", alert.ErrorCode)},
			},
		},
		{
			"type": "section",
			"fields": []map[string]string{
				{"type": "mrkdwn", "text": fmt.Sprintf("*Job ID:*\n%s", alert.JobID)},
				{"type": "mrkdwn", "text": fmt.Sprintf("*Date:*\n%s", alert.OccurredAt.Format(time.RFC3339))},
			},
		},
		{
			"type": "section",
			"text": map[string]string{
				"type": "mrkdwn",
				"text": fmt.Sprintf("*Error:*\n%s", errorMessage),
			},
		},
	}

	payload := map[string]any{
		"blocks": blocks,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal slack alert payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.alertWebhookURL, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create slack alert request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send slack alert: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.FromContext(ctx).Warn("slack alert failed", "status", resp.StatusCode)
		return fmt.Errorf("slack alert returned status %d", resp.StatusCode)
	}

	return nil
}

// SendRegistration は新規ユーザー登録通知を Slack に送信する
func (c *slackClient) SendRegistration(ctx context.Context, registration RegistrationNotification) error {
	if !c.IsRegistrationEnabled() {
		return nil
	}

	blocks := []map[string]any{
		{
			"type": "header",
			"text": map[string]string{
				"type": "plain_text",
				"text": "🎉 New User Registered",
			},
		},
		{
			"type": "section",
			"fields": []map[string]string{
				{"type": "mrkdwn", "text": fmt.Sprintf("*Name:*\n%s", registration.DisplayName)},
				{"type": "mrkdwn", "text": fmt.Sprintf("*Email:*\n%s", registration.Email)},
			},
		},
		{
			"type": "section",
			"fields": []map[string]string{
				{"type": "mrkdwn", "text": fmt.Sprintf("*Method:*\n%s", registration.Method)},
				{"type": "mrkdwn", "text": fmt.Sprintf("*Date:*\n%s", registration.CreatedAt.Format(time.RFC3339))},
			},
		},
		{
			"type": "section",
			"fields": []map[string]string{
				{"type": "mrkdwn", "text": fmt.Sprintf("*User ID:*\n%s", registration.UserID)},
			},
		},
	}

	payload := map[string]any{
		"blocks": blocks,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal slack registration payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.registrationWebhookURL, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create slack registration request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send slack registration notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.FromContext(ctx).Warn("slack registration notification failed", "status", resp.StatusCode)
		return fmt.Errorf("slack registration returned status %d", resp.StatusCode)
	}

	return nil
}
