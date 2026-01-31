package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
)

// ReactionRepository はリアクションのリポジトリインターフェースを表す
type ReactionRepository interface {
	FindLikesByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]model.Reaction, int64, error)
}

type reactionRepository struct {
	db *gorm.DB
}

// NewReactionRepository は reactionRepository を生成して ReactionRepository として返す
func NewReactionRepository(db *gorm.DB) ReactionRepository {
	return &reactionRepository{db: db}
}

// FindLikesByUserID はユーザーが高評価したエピソードのリアクション一覧を取得する
func (r *reactionRepository) FindLikesByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]model.Reaction, int64, error) {
	var reactions []model.Reaction
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Reaction{}).
		Where("user_id = ? AND reaction_type = ?", userID, model.ReactionTypeLike)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.
		Preload("Episode").
		Preload("Episode.Channel").
		Preload("Episode.Channel.Artwork").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&reactions).Error; err != nil {
		return nil, 0, err
	}

	return reactions, total, nil
}
