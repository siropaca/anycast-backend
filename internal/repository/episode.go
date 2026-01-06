package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/pkg/logger"
)

// エピソードデータへのアクセスインターフェース
type EpisodeRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*model.Episode, error)
	FindByChannelID(ctx context.Context, channelID uuid.UUID, filter EpisodeFilter) ([]model.Episode, int64, error)
	Create(ctx context.Context, episode *model.Episode) error
	Update(ctx context.Context, episode *model.Episode) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// エピソード検索のフィルタ条件
type EpisodeFilter struct {
	Status *string // "published" or "draft"
	Limit  int
	Offset int
}

type episodeRepository struct {
	db *gorm.DB
}

// EpisodeRepository の実装を返す
func NewEpisodeRepository(db *gorm.DB) EpisodeRepository {
	return &episodeRepository{db: db}
}

// 指定されたチャンネルのエピソード一覧を取得する
func (r *episodeRepository) FindByChannelID(ctx context.Context, channelID uuid.UUID, filter EpisodeFilter) ([]model.Episode, int64, error) {
	var episodes []model.Episode
	var total int64

	tx := r.db.WithContext(ctx).Model(&model.Episode{}).Where("channel_id = ?", channelID)

	// ステータスフィルタ
	if filter.Status != nil {
		switch *filter.Status {
		case "published":
			tx = tx.Where("published_at IS NOT NULL AND published_at <= ?", time.Now())
		case "draft":
			tx = tx.Where("published_at IS NULL")
		}
	}

	// 総件数を取得
	if err := tx.Count(&total).Error; err != nil {
		logger.FromContext(ctx).Error("failed to count episodes", "error", err, "channel_id", channelID)
		return nil, 0, apperror.ErrInternal.WithMessage("Failed to count episodes").WithError(err)
	}

	// ページネーションとリレーションのプリロード
	if err := tx.
		Preload("Artwork").
		Preload("FullAudio").
		Order("created_at DESC").
		Limit(filter.Limit).
		Offset(filter.Offset).
		Find(&episodes).Error; err != nil {
		logger.FromContext(ctx).Error("failed to fetch episodes", "error", err, "channel_id", channelID)
		return nil, 0, apperror.ErrInternal.WithMessage("Failed to fetch episodes").WithError(err)
	}

	return episodes, total, nil
}

// エピソードを作成する
func (r *episodeRepository) Create(ctx context.Context, episode *model.Episode) error {
	if err := r.db.WithContext(ctx).Create(episode).Error; err != nil {
		logger.FromContext(ctx).Error("failed to create episode", "error", err)
		return apperror.ErrInternal.WithMessage("Failed to create episode").WithError(err)
	}

	return nil
}

// 指定された ID のエピソードを取得する
func (r *episodeRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Episode, error) {
	var episode model.Episode

	if err := r.db.WithContext(ctx).
		Preload("Channel").
		Preload("Artwork").
		Preload("FullAudio").
		First(&episode, "id = ?", id).Error; err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrNotFound.WithMessage("Episode not found")
		}

		logger.FromContext(ctx).Error("failed to fetch episode", "error", err, "episode_id", id)
		return nil, apperror.ErrInternal.WithMessage("Failed to fetch episode").WithError(err)
	}

	return &episode, nil
}

// エピソードを更新する
func (r *episodeRepository) Update(ctx context.Context, episode *model.Episode) error {
	if err := r.db.WithContext(ctx).Save(episode).Error; err != nil {
		logger.FromContext(ctx).Error("failed to update episode", "error", err, "episode_id", episode.ID)
		return apperror.ErrInternal.WithMessage("Failed to update episode").WithError(err)
	}

	return nil
}

// エピソードを削除する
func (r *episodeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&model.Episode{}, "id = ?", id)
	if result.Error != nil {
		logger.FromContext(ctx).Error("failed to delete episode", "error", result.Error, "episode_id", id)
		return apperror.ErrInternal.WithMessage("Failed to delete episode").WithError(result.Error)
	}

	if result.RowsAffected == 0 {
		return apperror.ErrNotFound.WithMessage("Episode not found")
	}

	return nil
}
