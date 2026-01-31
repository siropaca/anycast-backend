package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
)

// FollowRepository はフォローのリポジトリインターフェースを表す
type FollowRepository interface {
	FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]model.Follow, int64, error)
	Create(ctx context.Context, follow *model.Follow) error
	DeleteByUserIDAndTargetUserID(ctx context.Context, userID, targetUserID uuid.UUID) error
	ExistsByUserIDAndTargetUserID(ctx context.Context, userID, targetUserID uuid.UUID) (bool, error)
}

type followRepository struct {
	db *gorm.DB
}

// NewFollowRepository は followRepository を生成して FollowRepository として返す
func NewFollowRepository(db *gorm.DB) FollowRepository {
	return &followRepository{db: db}
}

// FindByUserID はユーザーがフォロー中のユーザー一覧を取得する
func (r *followRepository) FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]model.Follow, int64, error) {
	var follows []model.Follow
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Follow{}).Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.
		Preload("TargetUser").
		Preload("TargetUser.Avatar").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&follows).Error; err != nil {
		return nil, 0, err
	}

	return follows, total, nil
}

// Create はフォローを作成する
func (r *followRepository) Create(ctx context.Context, follow *model.Follow) error {
	return r.db.WithContext(ctx).Create(follow).Error
}

// DeleteByUserIDAndTargetUserID はユーザーとターゲットユーザーに対応するフォローを削除する
func (r *followRepository) DeleteByUserIDAndTargetUserID(ctx context.Context, userID, targetUserID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Where("user_id = ? AND target_user_id = ?", userID, targetUserID).
		Delete(&model.Follow{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return apperror.ErrNotFound.WithMessage("フォローが見つかりません")
	}
	return nil
}

// ExistsByUserIDAndTargetUserID はフォローが存在するかどうかを返す
func (r *followRepository) ExistsByUserIDAndTargetUserID(ctx context.Context, userID, targetUserID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Follow{}).
		Where("user_id = ? AND target_user_id = ?", userID, targetUserID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
