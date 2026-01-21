package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/pkg/logger"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
)

// パスワード認証情報へのアクセスインターフェース
type CredentialRepository interface {
	Create(ctx context.Context, credential *model.Credential) error
	FindByUserID(ctx context.Context, userID uuid.UUID) (*model.Credential, error)
}

type credentialRepository struct {
	db *gorm.DB
}

// CredentialRepository の実装を返す
func NewCredentialRepository(db *gorm.DB) CredentialRepository {
	return &credentialRepository{db: db}
}

// パスワード認証情報を作成する
func (r *credentialRepository) Create(ctx context.Context, credential *model.Credential) error {
	if err := r.db.WithContext(ctx).Create(credential).Error; err != nil {
		logger.FromContext(ctx).Error("failed to create credential", "error", err)
		return apperror.ErrInternal.WithMessage("認証情報の作成に失敗しました").WithError(err)
	}

	return nil
}

// 指定されたユーザー ID のパスワード認証情報を取得する
func (r *credentialRepository) FindByUserID(ctx context.Context, userID uuid.UUID) (*model.Credential, error) {
	var credential model.Credential

	if err := r.db.WithContext(ctx).First(&credential, "user_id = ?", userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrNotFound.WithMessage("認証情報が見つかりません")
		}
		logger.FromContext(ctx).Error("failed to fetch credential", "error", err, "user_id", userID)
		return nil, apperror.ErrInternal.WithMessage("認証情報の取得に失敗しました").WithError(err)
	}

	return &credential, nil
}
