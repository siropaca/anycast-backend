package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/model"
)

// チャンネルデータへのアクセスインターフェース
type ChannelRepository interface {
	FindByUserID(ctx context.Context, userID uuid.UUID, filter ChannelFilter) ([]model.Channel, int64, error)
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
		return nil, 0, apperror.ErrInternal.WithMessage("Failed to count channels").WithError(err)
	}

	// ページネーションとリレーションのプリロード
	if err := tx.
		Preload("Category").
		Preload("Artwork").
		Order("created_at DESC").
		Limit(filter.Limit).
		Offset(filter.Offset).
		Find(&channels).Error; err != nil {
		return nil, 0, apperror.ErrInternal.WithMessage("Failed to fetch channels").WithError(err)
	}

	return channels, total, nil
}
