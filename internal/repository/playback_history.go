package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
)

// PlaybackHistoryRepository は再生履歴のリポジトリインターフェースを表す
type PlaybackHistoryRepository interface {
	FindByUserID(ctx context.Context, userID uuid.UUID, completed *bool, limit, offset int) ([]model.PlaybackHistory, int64, error)
	FindByUserIDAndEpisodeID(ctx context.Context, userID, episodeID uuid.UUID) (*model.PlaybackHistory, error)
	Upsert(ctx context.Context, history *model.PlaybackHistory) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type playbackHistoryRepository struct {
	db *gorm.DB
}

// NewPlaybackHistoryRepository は playbackHistoryRepository を生成して PlaybackHistoryRepository として返す
func NewPlaybackHistoryRepository(db *gorm.DB) PlaybackHistoryRepository {
	return &playbackHistoryRepository{db: db}
}

// FindByUserID はユーザーの再生履歴一覧を取得する
func (r *playbackHistoryRepository) FindByUserID(ctx context.Context, userID uuid.UUID, completed *bool, limit, offset int) ([]model.PlaybackHistory, int64, error) {
	var histories []model.PlaybackHistory
	var total int64

	query := r.db.WithContext(ctx).Model(&model.PlaybackHistory{}).Where("user_id = ?", userID)

	if completed != nil {
		query = query.Where("completed = ?", *completed)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.
		Preload("Episode").
		Preload("Episode.Channel").
		Preload("Episode.Channel.Artwork").
		Preload("Episode.FullAudio").
		Order("played_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&histories).Error; err != nil {
		return nil, 0, err
	}

	return histories, total, nil
}

// FindByUserIDAndEpisodeID はユーザーとエピソードに対応する再生履歴を取得する
func (r *playbackHistoryRepository) FindByUserIDAndEpisodeID(ctx context.Context, userID, episodeID uuid.UUID) (*model.PlaybackHistory, error) {
	var history model.PlaybackHistory
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND episode_id = ?", userID, episodeID).
		First(&history).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperror.ErrNotFound.WithMessage("再生履歴が見つかりません")
		}
		return nil, err
	}
	return &history, nil
}

// Upsert は再生履歴を作成または更新する
func (r *playbackHistoryRepository) Upsert(ctx context.Context, history *model.PlaybackHistory) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND episode_id = ?", history.UserID, history.EpisodeID).
		Assign(model.PlaybackHistory{
			ProgressMs: history.ProgressMs,
			Completed:  history.Completed,
			PlayedAt:   history.PlayedAt,
		}).
		FirstOrCreate(history).Error
}

// Delete は再生履歴を削除する
func (r *playbackHistoryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.PlaybackHistory{}, id).Error
}
