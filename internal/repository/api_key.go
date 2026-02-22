package repository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/pkg/logger"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
)

// APIKeyRepository は API キーへのアクセスインターフェース
type APIKeyRepository interface {
	Create(ctx context.Context, apiKey *model.APIKey) error
	FindByKeyHash(ctx context.Context, keyHash string) (*model.APIKey, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]model.APIKey, error)
	FindByUserIDAndID(ctx context.Context, userID, id uuid.UUID) (*model.APIKey, error)
	UpdateLastUsedAt(ctx context.Context, id uuid.UUID, lastUsedAt time.Time) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type apiKeyRepository struct {
	db *gorm.DB
}

// NewAPIKeyRepository は APIKeyRepository の実装を返す
func NewAPIKeyRepository(db *gorm.DB) APIKeyRepository {
	return &apiKeyRepository{db: db}
}

// Create は API キーを作成する
func (r *apiKeyRepository) Create(ctx context.Context, apiKey *model.APIKey) error {
	if err := r.db.WithContext(ctx).Create(apiKey).Error; err != nil {
		logger.FromContext(ctx).Error("failed to create api key", "error", err, "user_id", apiKey.UserID)
		return apperror.ErrInternal.WithMessage("API キーの作成に失敗しました").WithError(err)
	}

	return nil
}

// FindByKeyHash は指定されたハッシュの API キーを取得する
func (r *apiKeyRepository) FindByKeyHash(ctx context.Context, keyHash string) (*model.APIKey, error) {
	var apiKey model.APIKey

	if err := r.db.WithContext(ctx).First(&apiKey, "key_hash = ?", keyHash).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrNotFound.WithMessage("API キーが見つかりません")
		}

		logger.FromContext(ctx).Error("failed to fetch api key by hash", "error", err)
		return nil, apperror.ErrInternal.WithMessage("API キーの取得に失敗しました").WithError(err)
	}

	return &apiKey, nil
}

// FindByUserID は指定されたユーザーの全 API キーを取得する
func (r *apiKeyRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]model.APIKey, error) {
	var apiKeys []model.APIKey

	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Find(&apiKeys).Error; err != nil {
		logger.FromContext(ctx).Error("failed to fetch api keys", "error", err, "user_id", userID)
		return nil, apperror.ErrInternal.WithMessage("API キーの取得に失敗しました").WithError(err)
	}

	return apiKeys, nil
}

// FindByUserIDAndID は指定されたユーザーの指定された API キーを取得する
func (r *apiKeyRepository) FindByUserIDAndID(ctx context.Context, userID, id uuid.UUID) (*model.APIKey, error) {
	var apiKey model.APIKey

	if err := r.db.WithContext(ctx).First(&apiKey, "user_id = ? AND id = ?", userID, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrNotFound.WithMessage("API キーが見つかりません")
		}

		logger.FromContext(ctx).Error("failed to fetch api key", "error", err, "user_id", userID, "id", id)
		return nil, apperror.ErrInternal.WithMessage("API キーの取得に失敗しました").WithError(err)
	}

	return &apiKey, nil
}

// UpdateLastUsedAt は API キーの最終使用日時を更新する
func (r *apiKeyRepository) UpdateLastUsedAt(ctx context.Context, id uuid.UUID, lastUsedAt time.Time) error {
	if err := r.db.WithContext(ctx).Model(&model.APIKey{}).Where("id = ?", id).Update("last_used_at", lastUsedAt).Error; err != nil {
		logger.FromContext(ctx).Error("failed to update api key last_used_at", "error", err, "id", id)
		return apperror.ErrInternal.WithMessage("API キーの更新に失敗しました").WithError(err)
	}

	return nil
}

// Delete は API キーを削除する
func (r *apiKeyRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.APIKey{}).Error; err != nil {
		logger.FromContext(ctx).Error("failed to delete api key", "error", err, "id", id)
		return apperror.ErrInternal.WithMessage("API キーの削除に失敗しました").WithError(err)
	}

	return nil
}
