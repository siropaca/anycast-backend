package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
)

// ReactionRepository はリアクションのリポジトリインターフェースを表す
type ReactionRepository interface {
	FindLikesByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]model.Reaction, int64, error)
	FindByUserIDAndEpisodeID(ctx context.Context, userID, episodeID uuid.UUID) (*model.Reaction, error)
	Upsert(ctx context.Context, reaction *model.Reaction) (created bool, err error)
	DeleteByUserIDAndEpisodeID(ctx context.Context, userID, episodeID uuid.UUID) error
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

// FindByUserIDAndEpisodeID はユーザーとエピソードに対応するリアクションを取得する
func (r *reactionRepository) FindByUserIDAndEpisodeID(ctx context.Context, userID, episodeID uuid.UUID) (*model.Reaction, error) {
	var reaction model.Reaction
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND episode_id = ?", userID, episodeID).
		First(&reaction).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperror.ErrNotFound.WithMessage("リアクションが見つかりません")
		}
		return nil, err
	}
	return &reaction, nil
}

// Upsert はリアクションを作成または更新する
//
// @returns created - 新規作成の場合 true、更新の場合 false
func (r *reactionRepository) Upsert(ctx context.Context, reaction *model.Reaction) (bool, error) {
	var existing model.Reaction
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND episode_id = ?", reaction.UserID, reaction.EpisodeID).
		First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		if err := r.db.WithContext(ctx).Create(reaction).Error; err != nil {
			return false, err
		}
		return true, nil
	}
	if err != nil {
		return false, err
	}

	// 既存レコードを更新
	reaction.ID = existing.ID
	reaction.CreatedAt = existing.CreatedAt
	if err := r.db.WithContext(ctx).Model(&existing).Update("reaction_type", reaction.ReactionType).Error; err != nil {
		return false, err
	}
	return false, nil
}

// DeleteByUserIDAndEpisodeID はユーザーとエピソードに対応するリアクションを削除する
func (r *reactionRepository) DeleteByUserIDAndEpisodeID(ctx context.Context, userID, episodeID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Where("user_id = ? AND episode_id = ?", userID, episodeID).
		Delete(&model.Reaction{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return apperror.ErrNotFound.WithMessage("リアクションが見つかりません")
	}
	return nil
}
