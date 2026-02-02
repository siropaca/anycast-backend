package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/pkg/logger"
)

// ContactRepository はお問い合わせデータへのアクセスインターフェース
type ContactRepository interface {
	Create(ctx context.Context, contact *model.Contact) error
}

type contactRepository struct {
	db *gorm.DB
}

// NewContactRepository は ContactRepository の実装を返す
func NewContactRepository(db *gorm.DB) ContactRepository {
	return &contactRepository{db: db}
}

// Create はお問い合わせを作成する
func (r *contactRepository) Create(ctx context.Context, contact *model.Contact) error {
	if err := r.db.WithContext(ctx).Create(contact).Error; err != nil {
		logger.FromContext(ctx).Error("failed to create contact", "error", err)
		return apperror.ErrInternal.WithMessage("お問い合わせの作成に失敗しました").WithError(err)
	}

	return nil
}
