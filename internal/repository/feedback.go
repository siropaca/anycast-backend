package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/pkg/logger"
)

// FeedbackRepository はフィードバックデータへのアクセスインターフェース
type FeedbackRepository interface {
	Create(ctx context.Context, feedback *model.Feedback) error
}

type feedbackRepository struct {
	db *gorm.DB
}

// NewFeedbackRepository は FeedbackRepository の実装を返す
func NewFeedbackRepository(db *gorm.DB) FeedbackRepository {
	return &feedbackRepository{db: db}
}

// Create はフィードバックを作成する
func (r *feedbackRepository) Create(ctx context.Context, feedback *model.Feedback) error {
	if err := r.db.WithContext(ctx).Create(feedback).Error; err != nil {
		logger.FromContext(ctx).Error("failed to create feedback", "error", err)
		return apperror.ErrInternal.WithMessage("フィードバックの作成に失敗しました").WithError(err)
	}

	return nil
}
