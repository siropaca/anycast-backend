package service

import (
	"context"

	"github.com/siropaca/anycast-backend/internal/dto/response"
	"github.com/siropaca/anycast-backend/internal/infrastructure/slack"
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/pkg/logger"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
	"github.com/siropaca/anycast-backend/internal/repository"
)

// ContactService はお問い合わせ関連のビジネスロジックインターフェースを表す
type ContactService interface {
	CreateContact(ctx context.Context, input CreateContactInput) (*response.ContactDataResponse, error)
}

// CreateContactInput はお問い合わせ作成の入力を表す
type CreateContactInput struct {
	UserID    *string
	Category  string
	Email     string
	Name      string
	Content   string
	UserAgent *string
}

type contactService struct {
	contactRepo repository.ContactRepository
	slackClient slack.Client
}

// NewContactService は contactService を生成して ContactService として返す
func NewContactService(
	contactRepo repository.ContactRepository,
	slackClient slack.Client,
) ContactService {
	return &contactService{
		contactRepo: contactRepo,
		slackClient: slackClient,
	}
}

// CreateContact はお問い合わせを作成する
func (s *contactService) CreateContact(ctx context.Context, input CreateContactInput) (*response.ContactDataResponse, error) {
	category := model.ContactCategory(input.Category)

	contact := &model.Contact{
		Category:  category,
		Email:     input.Email,
		Name:      input.Name,
		Content:   input.Content,
		UserAgent: input.UserAgent,
	}

	// ログイン済みの場合は user_id をセット
	if input.UserID != nil {
		uid, err := uuid.Parse(*input.UserID)
		if err != nil {
			return nil, err
		}
		contact.UserID = &uid
	}

	// お問い合わせを保存
	if err := s.contactRepo.Create(ctx, contact); err != nil {
		return nil, err
	}

	// Slack 通知（非同期で実行し、エラーは無視）
	if s.slackClient.IsEnabled() {
		go func() {
			var userIDStr *string
			if contact.UserID != nil {
				s := contact.UserID.String()
				userIDStr = &s
			}
			notification := slack.ContactNotification{
				Category:      input.Category,
				CategoryLabel: category.Label(),
				Email:         input.Email,
				Name:          input.Name,
				Content:       input.Content,
				UserAgent:     input.UserAgent,
				UserID:        userIDStr,
				CreatedAt:     contact.CreatedAt,
			}
			if err := s.slackClient.SendContact(context.Background(), notification); err != nil {
				logger.Default().Warn("failed to send slack notification", "error", err)
			}
		}()
	}

	// レスポンスを構築
	resp := response.ContactResponse{
		ID:        contact.ID,
		Category:  input.Category,
		Email:     contact.Email,
		Name:      contact.Name,
		Content:   contact.Content,
		UserAgent: contact.UserAgent,
		CreatedAt: contact.CreatedAt,
	}

	return &response.ContactDataResponse{
		Data: resp,
	}, nil
}
