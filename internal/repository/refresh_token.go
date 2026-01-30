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

// RefreshTokenRepository はリフレッシュトークンへのアクセスインターフェース
type RefreshTokenRepository interface {
	Create(ctx context.Context, token *model.RefreshToken) error
	FindByToken(ctx context.Context, token string) (*model.RefreshToken, error)
	DeleteByToken(ctx context.Context, token string) error
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
}

type refreshTokenRepository struct {
	db *gorm.DB
}

// NewRefreshTokenRepository は RefreshTokenRepository の実装を返す
func NewRefreshTokenRepository(db *gorm.DB) RefreshTokenRepository {
	return &refreshTokenRepository{db: db}
}

// Create はリフレッシュトークンを作成する
func (r *refreshTokenRepository) Create(ctx context.Context, token *model.RefreshToken) error {
	if err := r.db.WithContext(ctx).Create(token).Error; err != nil {
		logger.FromContext(ctx).Error("failed to create refresh token", "error", err, "user_id", token.UserID)
		return apperror.ErrInternal.WithMessage("リフレッシュトークンの作成に失敗しました").WithError(err)
	}

	return nil
}

// FindByToken は指定されたトークン文字列のリフレッシュトークンを取得する
func (r *refreshTokenRepository) FindByToken(ctx context.Context, token string) (*model.RefreshToken, error) {
	var refreshToken model.RefreshToken

	if err := r.db.WithContext(ctx).First(&refreshToken, "token = ?", token).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrNotFound.WithMessage("リフレッシュトークンが見つかりません")
		}

		logger.FromContext(ctx).Error("failed to fetch refresh token", "error", err)
		return nil, apperror.ErrInternal.WithMessage("リフレッシュトークンの取得に失敗しました").WithError(err)
	}

	return &refreshToken, nil
}

// DeleteByToken は指定されたトークン文字列のリフレッシュトークンを削除する
func (r *refreshTokenRepository) DeleteByToken(ctx context.Context, token string) error {
	if err := r.db.WithContext(ctx).Where("token = ?", token).Delete(&model.RefreshToken{}).Error; err != nil {
		logger.FromContext(ctx).Error("failed to delete refresh token", "error", err)
		return apperror.ErrInternal.WithMessage("リフレッシュトークンの削除に失敗しました").WithError(err)
	}

	return nil
}

// DeleteByUserID は指定されたユーザーの全リフレッシュトークンを削除する
func (r *refreshTokenRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&model.RefreshToken{}).Error; err != nil {
		logger.FromContext(ctx).Error("failed to delete refresh tokens by user", "error", err, "user_id", userID)
		return apperror.ErrInternal.WithMessage("リフレッシュトークンの削除に失敗しました").WithError(err)
	}

	return nil
}
