package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/logger"
	"github.com/siropaca/anycast-backend/internal/model"
)

// チャンネルデータへのアクセスインターフェース
type ChannelRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*model.Channel, error)
	FindByUserID(ctx context.Context, userID uuid.UUID, filter ChannelFilter) ([]model.Channel, int64, error)
	Create(ctx context.Context, channel *model.Channel) error
	Update(ctx context.Context, channel *model.Channel) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// チャンネル検索のフィルタ条件
type ChannelFilter struct {
	Status *string // "published" or "draft"
	Limit  int
	Offset int
}

type channelRepository struct {
	db *gorm.DB
}

// ChannelRepository の実装を返す
func NewChannelRepository(db *gorm.DB) ChannelRepository {
	return &channelRepository{db: db}
}

// 指定されたユーザーのチャンネル一覧を取得する
func (r *channelRepository) FindByUserID(ctx context.Context, userID uuid.UUID, filter ChannelFilter) ([]model.Channel, int64, error) {
	var channels []model.Channel
	var total int64

	tx := r.db.WithContext(ctx).Model(&model.Channel{}).Where("user_id = ?", userID)

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
		logger.FromContext(ctx).Error("failed to count channels", "error", err, "user_id", userID)
		return nil, 0, apperror.ErrInternal.WithMessage("Failed to count channels").WithError(err)
	}

	// ページネーションとリレーションのプリロード
	if err := tx.
		Preload("Category").
		Preload("Artwork").
		Preload("Characters").
		Preload("Characters.Voice").
		Order("created_at DESC").
		Limit(filter.Limit).
		Offset(filter.Offset).
		Find(&channels).Error; err != nil {
		logger.FromContext(ctx).Error("failed to fetch channels", "error", err, "user_id", userID)
		return nil, 0, apperror.ErrInternal.WithMessage("Failed to fetch channels").WithError(err)
	}

	return channels, total, nil
}

// 指定された ID のチャンネルを取得する
func (r *channelRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Channel, error) {
	var channel model.Channel
	if err := r.db.WithContext(ctx).
		Preload("Category").
		Preload("Artwork").
		Preload("Characters").
		Preload("Characters.Voice").
		First(&channel, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrNotFound.WithMessage("Channel not found")
		}
		logger.FromContext(ctx).Error("failed to fetch channel", "error", err, "channel_id", id)
		return nil, apperror.ErrInternal.WithMessage("Failed to fetch channel").WithError(err)
	}
	return &channel, nil
}

// チャンネルを作成する
func (r *channelRepository) Create(ctx context.Context, channel *model.Channel) error {
	if err := r.db.WithContext(ctx).Create(channel).Error; err != nil {
		logger.FromContext(ctx).Error("failed to create channel", "error", err)
		return apperror.ErrInternal.WithMessage("Failed to create channel").WithError(err)
	}
	return nil
}

// チャンネルを更新する
func (r *channelRepository) Update(ctx context.Context, channel *model.Channel) error {
	if err := r.db.WithContext(ctx).Save(channel).Error; err != nil {
		logger.FromContext(ctx).Error("failed to update channel", "error", err, "channel_id", channel.ID)
		return apperror.ErrInternal.WithMessage("Failed to update channel").WithError(err)
	}
	return nil
}

// チャンネルを削除する
func (r *channelRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&model.Channel{}, "id = ?", id)
	if result.Error != nil {
		logger.FromContext(ctx).Error("failed to delete channel", "error", result.Error, "channel_id", id)
		return apperror.ErrInternal.WithMessage("Failed to delete channel").WithError(result.Error)
	}
	if result.RowsAffected == 0 {
		return apperror.ErrNotFound.WithMessage("Channel not found")
	}
	return nil
}
