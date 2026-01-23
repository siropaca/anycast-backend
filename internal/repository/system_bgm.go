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

// SystemBgmRepository はシステム BGM データへのアクセスインターフェース
type SystemBgmRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*model.SystemBgm, error)
	FindActive(ctx context.Context, filter SystemBgmFilter) ([]model.SystemBgm, int64, error)
	CountActive(ctx context.Context) (int64, error)
}

// SystemBgmFilter はシステム BGM 検索のフィルタ条件を表す
type SystemBgmFilter struct {
	Limit  int
	Offset int
}

type systemBgmRepository struct {
	db *gorm.DB
}

// NewSystemBgmRepository は SystemBgmRepository の実装を返す
func NewSystemBgmRepository(db *gorm.DB) SystemBgmRepository {
	return &systemBgmRepository{db: db}
}

// FindActive はアクティブなシステム BGM 一覧を取得する
func (r *systemBgmRepository) FindActive(ctx context.Context, filter SystemBgmFilter) ([]model.SystemBgm, int64, error) {
	var bgms []model.SystemBgm
	var total int64

	tx := r.db.WithContext(ctx).Model(&model.SystemBgm{}).Where("is_active = ?", true)

	// 総件数を取得
	if err := tx.Count(&total).Error; err != nil {
		logger.FromContext(ctx).Error("failed to count system bgms", "error", err)
		return nil, 0, apperror.ErrInternal.WithMessage("システム BGM 数の取得に失敗しました").WithError(err)
	}

	// ページネーションとリレーションのプリロード
	if err := tx.
		Preload("Audio").
		Order("sort_order ASC").
		Limit(filter.Limit).
		Offset(filter.Offset).
		Find(&bgms).Error; err != nil {
		logger.FromContext(ctx).Error("failed to fetch system bgms", "error", err)
		return nil, 0, apperror.ErrInternal.WithMessage("システム BGM 一覧の取得に失敗しました").WithError(err)
	}

	return bgms, total, nil
}

// CountActive はアクティブなシステム BGM の総件数を取得する
func (r *systemBgmRepository) CountActive(ctx context.Context) (int64, error) {
	var total int64

	if err := r.db.WithContext(ctx).
		Model(&model.SystemBgm{}).
		Where("is_active = ?", true).
		Count(&total).Error; err != nil {
		logger.FromContext(ctx).Error("failed to count active system bgms", "error", err)
		return 0, apperror.ErrInternal.WithMessage("システム BGM 数の取得に失敗しました").WithError(err)
	}

	return total, nil
}

// FindByID は指定された ID のシステム BGM を取得する
func (r *systemBgmRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.SystemBgm, error) {
	var bgm model.SystemBgm

	if err := r.db.WithContext(ctx).
		Preload("Audio").
		First(&bgm, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrNotFound.WithMessage("システム BGM が見つかりません")
		}

		logger.FromContext(ctx).Error("failed to fetch system bgm", "error", err, "bgm_id", id)
		return nil, apperror.ErrInternal.WithMessage("システム BGM の取得に失敗しました").WithError(err)
	}

	return &bgm, nil
}
