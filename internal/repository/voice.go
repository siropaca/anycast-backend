package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/pkg/logger"
)

// VoiceRepository はボイスデータへのアクセスインターフェース
type VoiceRepository interface {
	FindAll(ctx context.Context, filter VoiceFilter) ([]model.Voice, error)
	FindByID(ctx context.Context, id string) (*model.Voice, error)
	FindActiveByID(ctx context.Context, id string) (*model.Voice, error)
}

// VoiceFilter はボイス検索のフィルタ条件を表す
type VoiceFilter struct {
	Provider *string
	Gender   *string
}

type voiceRepository struct {
	db *gorm.DB
}

// NewVoiceRepository は VoiceRepository の実装を返す
func NewVoiceRepository(db *gorm.DB) VoiceRepository {
	return &voiceRepository{db: db}
}

// FindAll はフィルタ条件に基づいてアクティブなボイス一覧を取得する
func (r *voiceRepository) FindAll(ctx context.Context, filter VoiceFilter) ([]model.Voice, error) {
	var voices []model.Voice
	tx := r.db.WithContext(ctx).Model(&model.Voice{}).
		Where("is_active = ?", true)

	if filter.Provider != nil {
		tx = tx.Where("provider = ?", *filter.Provider)
	}
	if filter.Gender != nil {
		tx = tx.Where("gender = ?", *filter.Gender)
	}

	if err := tx.Find(&voices).Error; err != nil {
		logger.FromContext(ctx).Error("failed to fetch voices", "error", err)
		return nil, apperror.ErrInternal.WithMessage("ボイス一覧の取得に失敗しました").WithError(err)
	}

	return voices, nil
}

// FindByID は指定された ID のボイスを取得する
func (r *voiceRepository) FindByID(ctx context.Context, id string) (*model.Voice, error) {
	var voice model.Voice

	if err := r.db.WithContext(ctx).First(&voice, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrNotFound.WithMessage("ボイスが見つかりません")
		}
		logger.FromContext(ctx).Error("failed to fetch voice", "error", err, "voice_id", id)
		return nil, apperror.ErrInternal.WithMessage("ボイスの取得に失敗しました").WithError(err)
	}

	return &voice, nil
}

// FindActiveByID は指定された ID のアクティブなボイスを取得する
func (r *voiceRepository) FindActiveByID(ctx context.Context, id string) (*model.Voice, error) {
	var voice model.Voice

	if err := r.db.WithContext(ctx).Where("is_active = ?", true).First(&voice, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrNotFound.WithMessage("ボイスが見つからないか、無効です")
		}
		logger.FromContext(ctx).Error("failed to fetch active voice", "error", err, "voice_id", id)
		return nil, apperror.ErrInternal.WithMessage("ボイスの取得に失敗しました").WithError(err)
	}

	return &voice, nil
}
