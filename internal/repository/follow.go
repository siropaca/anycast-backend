package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
)

// FollowRepository はフォローのリポジトリインターフェースを表す
type FollowRepository interface {
	FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]model.Follow, int64, error)
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
