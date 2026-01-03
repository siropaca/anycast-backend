package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/logger"
	"github.com/siropaca/anycast-backend/internal/model"
)

// エピソードデータへのアクセスインターフェース
type EpisodeRepository interface {
	FindByChannelID(ctx context.Context, channelID uuid.UUID, filter EpisodeFilter) ([]model.Episode, int64, error)
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
