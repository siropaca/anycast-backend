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

// BGM データへのアクセスインターフェース
type BgmRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*model.Bgm, error)
	FindByUserID(ctx context.Context, userID uuid.UUID, filter BgmFilter) ([]model.Bgm, int64, error)
	Create(ctx context.Context, bgm *model.Bgm) error
	Update(ctx context.Context, bgm *model.Bgm) error
	Delete(ctx context.Context, id uuid.UUID) error
	IsUsedInAnyEpisode(ctx context.Context, id uuid.UUID) (bool, error)
	ExistsByUserIDAndName(ctx context.Context, userID uuid.UUID, name string, excludeID *uuid.UUID) (bool, error)
}

// BGM 検索のフィルタ条件
type BgmFilter struct {
	Limit  int
	Offset int
}

type bgmRepository struct {
	db *gorm.DB
}

// BgmRepository の実装を返す
func NewBgmRepository(db *gorm.DB) BgmRepository {
	return &bgmRepository{db: db}
}

// 指定されたユーザーの BGM 一覧を取得する
func (r *bgmRepository) FindByUserID(ctx context.Context, userID uuid.UUID, filter BgmFilter) ([]model.Bgm, int64, error) {
	var bgms []model.Bgm
	var total int64

	tx := r.db.WithContext(ctx).Model(&model.Bgm{}).Where("user_id = ?", userID)

	// 総件数を取得
	if err := tx.Count(&total).Error; err != nil {
		logger.FromContext(ctx).Error("failed to count bgms", "error", err, "user_id", userID)
		return nil, 0, apperror.ErrInternal.WithMessage("BGM 数の取得に失敗しました").WithError(err)
	}

	// ページネーションとリレーションのプリロード
	if err := tx.
		Preload("Audio").
		Preload("Episodes.Channel").
		Preload("Channels").
		Order("created_at DESC").
		Limit(filter.Limit).
		Offset(filter.Offset).
		Find(&bgms).Error; err != nil {
		logger.FromContext(ctx).Error("failed to fetch bgms", "error", err, "user_id", userID)
		return nil, 0, apperror.ErrInternal.WithMessage("BGM 一覧の取得に失敗しました").WithError(err)
	}

	return bgms, total, nil
}

// 指定された ID の BGM を取得する
func (r *bgmRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Bgm, error) {
	var bgm model.Bgm

	if err := r.db.WithContext(ctx).
		Preload("Audio").
		Preload("Episodes.Channel").
		Preload("Channels").
		First(&bgm, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrNotFound.WithMessage("BGM が見つかりません")
		}

		logger.FromContext(ctx).Error("failed to fetch bgm", "error", err, "bgm_id", id)
		return nil, apperror.ErrInternal.WithMessage("BGM の取得に失敗しました").WithError(err)
	}

	return &bgm, nil
}

// BGM を作成する
func (r *bgmRepository) Create(ctx context.Context, bgm *model.Bgm) error {
	if err := r.db.WithContext(ctx).Create(bgm).Error; err != nil {
		logger.FromContext(ctx).Error("failed to create bgm", "error", err)
		return apperror.ErrInternal.WithMessage("BGM の作成に失敗しました").WithError(err)
	}

	return nil
}

// BGM を更新する
func (r *bgmRepository) Update(ctx context.Context, bgm *model.Bgm) error {
	if err := r.db.WithContext(ctx).Save(bgm).Error; err != nil {
		logger.FromContext(ctx).Error("failed to update bgm", "error", err, "bgm_id", bgm.ID)
		return apperror.ErrInternal.WithMessage("BGM の更新に失敗しました").WithError(err)
	}

	return nil
}

// BGM を削除する
func (r *bgmRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&model.Bgm{}, "id = ?", id)
	if result.Error != nil {
		logger.FromContext(ctx).Error("failed to delete bgm", "error", result.Error, "bgm_id", id)
		return apperror.ErrInternal.WithMessage("BGM の削除に失敗しました").WithError(result.Error)
	}
	if result.RowsAffected == 0 {
		return apperror.ErrNotFound.WithMessage("BGM が見つかりません")
	}

	return nil
}

// BGM がいずれかのエピソードで使用中かどうかを確認する
func (r *bgmRepository) IsUsedInAnyEpisode(ctx context.Context, id uuid.UUID) (bool, error) {
	var count int64

	if err := r.db.WithContext(ctx).
		Model(&model.Episode{}).
		Where("bgm_id = ?", id).
		Count(&count).Error; err != nil {
		logger.FromContext(ctx).Error("failed to check bgm usage", "error", err, "bgm_id", id)
		return false, apperror.ErrInternal.WithMessage("BGM の使用状況確認に失敗しました").WithError(err)
	}

	return count > 0, nil
}

// 同一ユーザー内で同じ名前の BGM が存在するかどうかを確認する
func (r *bgmRepository) ExistsByUserIDAndName(ctx context.Context, userID uuid.UUID, name string, excludeID *uuid.UUID) (bool, error) {
	var count int64

	tx := r.db.WithContext(ctx).
		Model(&model.Bgm{}).
		Where("user_id = ? AND name = ?", userID, name)

	if excludeID != nil {
		tx = tx.Where("id != ?", *excludeID)
	}

	if err := tx.Count(&count).Error; err != nil {
		logger.FromContext(ctx).Error("failed to check bgm name", "error", err, "user_id", userID, "name", name)
		return false, apperror.ErrInternal.WithMessage("BGM 名の確認に失敗しました").WithError(err)
	}

	return count > 0, nil
}
