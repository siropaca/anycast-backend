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

// ChannelRepository はチャンネルデータへのアクセスインターフェース
type ChannelRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*model.Channel, error)
	FindByUserID(ctx context.Context, userID uuid.UUID, filter ChannelFilter) ([]model.Channel, int64, error)
	FindPublishedByUserID(ctx context.Context, userID uuid.UUID) ([]model.Channel, error)
	Search(ctx context.Context, filter SearchChannelFilter) ([]model.Channel, int64, error)
	Create(ctx context.Context, channel *model.Channel) error
	Update(ctx context.Context, channel *model.Channel) error
	Delete(ctx context.Context, id uuid.UUID) error
	ReplaceChannelCharacters(ctx context.Context, channelID uuid.UUID, characterIDs []uuid.UUID) error
}

// ChannelFilter はチャンネル検索のフィルタ条件を表す
type ChannelFilter struct {
	Status *string // "published" or "draft"
	Limit  int
	Offset int
}

// SearchChannelFilter はチャンネル検索のフィルタ条件を表す
type SearchChannelFilter struct {
	Query        string
	CategorySlug *string
	Limit        int
	Offset       int
}

type channelRepository struct {
	db *gorm.DB
}

// NewChannelRepository は ChannelRepository の実装を返す
func NewChannelRepository(db *gorm.DB) ChannelRepository {
	return &channelRepository{db: db}
}

// FindByUserID は指定されたユーザーのチャンネル一覧を取得する
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
		return nil, 0, apperror.ErrInternal.WithMessage("チャンネル数の取得に失敗しました").WithError(err)
	}

	// ページネーションとリレーションのプリロード
	if err := tx.
		Preload("User").
		Preload("User.Avatar").
		Preload("Category").
		Preload("Artwork").
		Preload("DefaultBgm").
		Preload("DefaultBgm.Audio").
		Preload("DefaultSystemBgm").
		Preload("DefaultSystemBgm.Audio").
		Preload("ChannelCharacters").
		Preload("ChannelCharacters.Character").
		Preload("ChannelCharacters.Character.Avatar").
		Preload("ChannelCharacters.Character.Voice").
		Order("created_at DESC").
		Limit(filter.Limit).
		Offset(filter.Offset).
		Find(&channels).Error; err != nil {
		logger.FromContext(ctx).Error("failed to fetch channels", "error", err, "user_id", userID)
		return nil, 0, apperror.ErrInternal.WithMessage("チャンネル一覧の取得に失敗しました").WithError(err)
	}

	return channels, total, nil
}

// FindPublishedByUserID は指定されたユーザーの公開済みチャンネル一覧を取得する
func (r *channelRepository) FindPublishedByUserID(ctx context.Context, userID uuid.UUID) ([]model.Channel, error) {
	var channels []model.Channel

	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND published_at IS NOT NULL AND published_at <= ?", userID, time.Now()).
		Preload("Category").
		Preload("Artwork").
		Order("created_at DESC").
		Find(&channels).Error; err != nil {
		logger.FromContext(ctx).Error("failed to fetch published channels", "error", err, "user_id", userID)
		return nil, apperror.ErrInternal.WithMessage("公開チャンネル一覧の取得に失敗しました").WithError(err)
	}

	return channels, nil
}

// FindByID は指定された ID のチャンネルを取得する
func (r *channelRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Channel, error) {
	var channel model.Channel

	if err := r.db.WithContext(ctx).
		Preload("User").
		Preload("User.Avatar").
		Preload("Category").
		Preload("Artwork").
		Preload("DefaultBgm").
		Preload("DefaultBgm.Audio").
		Preload("DefaultSystemBgm").
		Preload("DefaultSystemBgm.Audio").
		Preload("ChannelCharacters").
		Preload("ChannelCharacters.Character").
		Preload("ChannelCharacters.Character.Avatar").
		Preload("ChannelCharacters.Character.Voice").
		First(&channel, "id = ?", id).Error; err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrNotFound.WithMessage("チャンネルが見つかりません")
		}

		logger.FromContext(ctx).Error("failed to fetch channel", "error", err, "channel_id", id)
		return nil, apperror.ErrInternal.WithMessage("チャンネルの取得に失敗しました").WithError(err)
	}

	return &channel, nil
}

// Create はチャンネルを作成する
func (r *channelRepository) Create(ctx context.Context, channel *model.Channel) error {
	if err := r.db.WithContext(ctx).Create(channel).Error; err != nil {
		logger.FromContext(ctx).Error("failed to create channel", "error", err)
		return apperror.ErrInternal.WithMessage("チャンネルの作成に失敗しました").WithError(err)
	}

	return nil
}

// Update はチャンネルを更新する
func (r *channelRepository) Update(ctx context.Context, channel *model.Channel) error {
	if err := r.db.WithContext(ctx).Save(channel).Error; err != nil {
		logger.FromContext(ctx).Error("failed to update channel", "error", err, "channel_id", channel.ID)
		return apperror.ErrInternal.WithMessage("チャンネルの更新に失敗しました").WithError(err)
	}

	return nil
}

// Delete はチャンネルを削除する
func (r *channelRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&model.Channel{}, "id = ?", id)
	if result.Error != nil {
		logger.FromContext(ctx).Error("failed to delete channel", "error", result.Error, "channel_id", id)
		return apperror.ErrInternal.WithMessage("チャンネルの削除に失敗しました").WithError(result.Error)
	}
	if result.RowsAffected == 0 {
		return apperror.ErrNotFound.WithMessage("チャンネルが見つかりません")
	}

	return nil
}

// Search は公開チャンネルをキーワードで検索する
func (r *channelRepository) Search(ctx context.Context, filter SearchChannelFilter) ([]model.Channel, int64, error) {
	var channels []model.Channel
	var total int64

	keyword := "%" + filter.Query + "%"

	tx := r.db.WithContext(ctx).Model(&model.Channel{}).
		Where("published_at IS NOT NULL AND published_at <= ?", time.Now()).
		Where("(name ILIKE ? OR description ILIKE ?)", keyword, keyword)

	// カテゴリスラッグでフィルタ
	if filter.CategorySlug != nil {
		tx = tx.Joins("JOIN categories ON categories.id = channels.category_id").
			Where("categories.slug = ?", *filter.CategorySlug)
	}

	// 総件数を取得
	if err := tx.Count(&total).Error; err != nil {
		logger.FromContext(ctx).Error("failed to count search results", "error", err)
		return nil, 0, apperror.ErrInternal.WithMessage("チャンネル検索の件数取得に失敗しました").WithError(err)
	}

	// ページネーションとリレーションのプリロード
	if err := tx.
		Preload("Category").
		Preload("Artwork").
		Order("created_at DESC").
		Limit(filter.Limit).
		Offset(filter.Offset).
		Find(&channels).Error; err != nil {
		logger.FromContext(ctx).Error("failed to search channels", "error", err)
		return nil, 0, apperror.ErrInternal.WithMessage("チャンネル検索に失敗しました").WithError(err)
	}

	return channels, total, nil
}

// ReplaceChannelCharacters はチャンネルに紐づくキャラクターを置き換える
func (r *channelRepository) ReplaceChannelCharacters(ctx context.Context, channelID uuid.UUID, characterIDs []uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 既存の紐づけを削除
		if err := tx.Where("channel_id = ?", channelID).Delete(&model.ChannelCharacter{}).Error; err != nil {
			logger.FromContext(ctx).Error("failed to delete channel characters", "error", err, "channel_id", channelID)
			return apperror.ErrInternal.WithMessage("チャンネルのキャラクター更新に失敗しました").WithError(err)
		}

		// 新しい紐づけを作成
		for _, characterID := range characterIDs {
			channelCharacter := model.ChannelCharacter{
				ChannelID:   channelID,
				CharacterID: characterID,
			}
			if err := tx.Create(&channelCharacter).Error; err != nil {
				logger.FromContext(ctx).Error("failed to create channel character", "error", err, "channel_id", channelID, "character_id", characterID)
				return apperror.ErrInternal.WithMessage("チャンネルのキャラクター更新に失敗しました").WithError(err)
			}
		}

		return nil
	})
}
