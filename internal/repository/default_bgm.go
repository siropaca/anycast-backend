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

// デフォルト BGM データへのアクセスインターフェース
type DefaultBgmRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*model.DefaultBgm, error)
	FindActive(ctx context.Context, filter DefaultBgmFilter) ([]model.DefaultBgm, int64, error)
	CountActive(ctx context.Context) (int64, error)
}

// デフォルト BGM 検索のフィルタ条件
type DefaultBgmFilter struct {
	Limit  int
	Offset int
}

type defaultBgmRepository struct {
	db *gorm.DB
}

// DefaultBgmRepository の実装を返す
func NewDefaultBgmRepository(db *gorm.DB) DefaultBgmRepository {
	return &defaultBgmRepository{db: db}
}

// アクティブなデフォルト BGM 一覧を取得する
func (r *defaultBgmRepository) FindActive(ctx context.Context, filter DefaultBgmFilter) ([]model.DefaultBgm, int64, error) {
	var bgms []model.DefaultBgm
	var total int64

	tx := r.db.WithContext(ctx).Model(&model.DefaultBgm{}).Where("is_active = ?", true)

	// 総件数を取得
	if err := tx.Count(&total).Error; err != nil {
		logger.FromContext(ctx).Error("failed to count default bgms", "error", err)
		return nil, 0, apperror.ErrInternal.WithMessage("デフォルト BGM 数の取得に失敗しました").WithError(err)
	}

	// ページネーションとリレーションのプリロード
	if err := tx.
		Preload("Audio").
		Order("sort_order ASC").
		Limit(filter.Limit).
		Offset(filter.Offset).
		Find(&bgms).Error; err != nil {
		logger.FromContext(ctx).Error("failed to fetch default bgms", "error", err)
		return nil, 0, apperror.ErrInternal.WithMessage("デフォルト BGM 一覧の取得に失敗しました").WithError(err)
	}

	return bgms, total, nil
}

// アクティブなデフォルト BGM の総件数を取得する
func (r *defaultBgmRepository) CountActive(ctx context.Context) (int64, error) {
	var total int64

	if err := r.db.WithContext(ctx).
		Model(&model.DefaultBgm{}).
		Where("is_active = ?", true).
		Count(&total).Error; err != nil {
		logger.FromContext(ctx).Error("failed to count active default bgms", "error", err)
		return 0, apperror.ErrInternal.WithMessage("デフォルト BGM 数の取得に失敗しました").WithError(err)
	}

	return total, nil
}

// 指定された ID のデフォルト BGM を取得する
func (r *defaultBgmRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.DefaultBgm, error) {
	var bgm model.DefaultBgm

	if err := r.db.WithContext(ctx).
		Preload("Audio").
		First(&bgm, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrNotFound.WithMessage("デフォルト BGM が見つかりません")
		}

		logger.FromContext(ctx).Error("failed to fetch default bgm", "error", err, "bgm_id", id)
		return nil, apperror.ErrInternal.WithMessage("デフォルト BGM の取得に失敗しました").WithError(err)
	}

	return &bgm, nil
}
