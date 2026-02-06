package repository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/pkg/logger"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
)

// EpisodeRepository はエピソードデータへのアクセスインターフェース
type EpisodeRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*model.Episode, error)
	FindByChannelID(ctx context.Context, channelID uuid.UUID, filter EpisodeFilter) ([]model.Episode, int64, error)
	CountPublishedByChannelIDs(ctx context.Context, channelIDs []uuid.UUID) (map[uuid.UUID]int, error)
	Create(ctx context.Context, episode *model.Episode) error
	Update(ctx context.Context, episode *model.Episode) error
	Delete(ctx context.Context, id uuid.UUID) error
	IncrementPlayCount(ctx context.Context, id uuid.UUID) error
}

// EpisodeFilter はエピソード検索のフィルタ条件を表す
type EpisodeFilter struct {
	Status *string // "published" or "draft"
	Limit  int
	Offset int
}

type episodeRepository struct {
	db *gorm.DB
}

// NewEpisodeRepository は EpisodeRepository の実装を返す
func NewEpisodeRepository(db *gorm.DB) EpisodeRepository {
	return &episodeRepository{db: db}
}

// FindByChannelID は指定されたチャンネルのエピソード一覧を取得する
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
		return nil, 0, apperror.ErrInternal.WithMessage("エピソード数の取得に失敗しました").WithError(err)
	}

	// ページネーションとリレーションのプリロード
	if err := tx.
		Preload("Artwork").
		Preload("FullAudio").
		Preload("Bgm").
		Preload("Bgm.Audio").
		Preload("SystemBgm").
		Preload("SystemBgm.Audio").
		Order("created_at DESC").
		Limit(filter.Limit).
		Offset(filter.Offset).
		Find(&episodes).Error; err != nil {
		logger.FromContext(ctx).Error("failed to fetch episodes", "error", err, "channel_id", channelID)
		return nil, 0, apperror.ErrInternal.WithMessage("エピソード一覧の取得に失敗しました").WithError(err)
	}

	return episodes, total, nil
}

// CountPublishedByChannelIDs は指定されたチャンネル群ごとの公開済みエピソード数を取得する
func (r *episodeRepository) CountPublishedByChannelIDs(ctx context.Context, channelIDs []uuid.UUID) (map[uuid.UUID]int, error) {
	if len(channelIDs) == 0 {
		return map[uuid.UUID]int{}, nil
	}

	type countRow struct {
		ChannelID uuid.UUID
		Count     int
	}

	var rows []countRow
	if err := r.db.WithContext(ctx).
		Model(&model.Episode{}).
		Select("channel_id, COUNT(*) AS count").
		Where("channel_id IN ? AND published_at IS NOT NULL AND published_at <= ?", channelIDs, time.Now()).
		Group("channel_id").
		Find(&rows).Error; err != nil {
		logger.FromContext(ctx).Error("failed to count published episodes by channel IDs", "error", err)
		return nil, apperror.ErrInternal.WithMessage("エピソード数の取得に失敗しました").WithError(err)
	}

	result := make(map[uuid.UUID]int, len(rows))
	for _, row := range rows {
		result[row.ChannelID] = row.Count
	}

	return result, nil
}

// Create はエピソードを作成する
func (r *episodeRepository) Create(ctx context.Context, episode *model.Episode) error {
	if err := r.db.WithContext(ctx).Create(episode).Error; err != nil {
		logger.FromContext(ctx).Error("failed to create episode", "error", err)
		return apperror.ErrInternal.WithMessage("エピソードの作成に失敗しました").WithError(err)
	}

	return nil
}

// FindByID は指定された ID のエピソードを取得する
func (r *episodeRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Episode, error) {
	var episode model.Episode

	if err := r.db.WithContext(ctx).
		Preload("Channel").
		Preload("Artwork").
		Preload("FullAudio").
		Preload("Bgm").
		Preload("Bgm.Audio").
		Preload("SystemBgm").
		Preload("SystemBgm.Audio").
		First(&episode, "id = ?", id).Error; err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrNotFound.WithMessage("エピソードが見つかりません")
		}

		logger.FromContext(ctx).Error("failed to fetch episode", "error", err, "episode_id", id)
		return nil, apperror.ErrInternal.WithMessage("エピソードの取得に失敗しました").WithError(err)
	}

	return &episode, nil
}

// Update はエピソードを更新する
func (r *episodeRepository) Update(ctx context.Context, episode *model.Episode) error {
	if err := r.db.WithContext(ctx).Save(episode).Error; err != nil {
		logger.FromContext(ctx).Error("failed to update episode", "error", err, "episode_id", episode.ID)
		return apperror.ErrInternal.WithMessage("エピソードの更新に失敗しました").WithError(err)
	}

	return nil
}

// Delete はエピソードを削除する
func (r *episodeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&model.Episode{}, "id = ?", id)
	if result.Error != nil {
		logger.FromContext(ctx).Error("failed to delete episode", "error", result.Error, "episode_id", id)
		return apperror.ErrInternal.WithMessage("エピソードの削除に失敗しました").WithError(result.Error)
	}

	if result.RowsAffected == 0 {
		return apperror.ErrNotFound.WithMessage("エピソードが見つかりません")
	}

	return nil
}

// IncrementPlayCount は指定されたエピソードの再生回数をアトミックにインクリメントする
func (r *episodeRepository) IncrementPlayCount(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Model(&model.Episode{}).
		Where("id = ?", id).
		UpdateColumn("play_count", gorm.Expr("play_count + 1"))
	if result.Error != nil {
		logger.FromContext(ctx).Error("failed to increment play count", "error", result.Error, "episode_id", id)
		return apperror.ErrInternal.WithMessage("再生回数の更新に失敗しました").WithError(result.Error)
	}

	if result.RowsAffected == 0 {
		return apperror.ErrNotFound.WithMessage("エピソードが見つかりません")
	}

	return nil
}
