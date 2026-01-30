package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
)

// RecommendedChannel はおすすめチャンネルの集計データを保持する
type RecommendedChannel struct {
	model.Channel
	EpisodeCount    int        `gorm:"column:episode_count"`
	TotalPlayCount  int        `gorm:"column:total_play_count"`
	LatestEpisodeAt *time.Time `gorm:"column:latest_episode_at"`
}

// CategoryPreference はユーザーのカテゴリ傾向を表す
type CategoryPreference struct {
	CategoryID uuid.UUID `gorm:"column:category_id"`
	PlayCount  int       `gorm:"column:play_count"`
}

// RecommendChannelParams はおすすめチャンネル取得のパラメータ
type RecommendChannelParams struct {
	CategoryID *uuid.UUID
	Limit      int
	Offset     int
}

// RecommendEpisodeParams はおすすめエピソード取得のパラメータ
type RecommendEpisodeParams struct {
	CategoryID *uuid.UUID
	Limit      int
	Offset     int
}

// RecommendationRepository はおすすめ取得用のリポジトリインターフェースを表す
type RecommendationRepository interface {
	// チャンネル関連
	FindRecommendedChannels(ctx context.Context, params RecommendChannelParams) ([]RecommendedChannel, int64, error)
	FindUserCategoryPreferences(ctx context.Context, userID uuid.UUID) ([]CategoryPreference, error)
	FindUserPlayedChannelIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)
	FindUserDefaultPlaylistCategoryPreferences(ctx context.Context, userID uuid.UUID) ([]CategoryPreference, error)
	FindUserChannelIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)

	// エピソード関連
	FindPublishedEpisodes(ctx context.Context, params RecommendEpisodeParams) ([]model.Episode, int64, error)
	FindUserPlaybackHistories(ctx context.Context, userID uuid.UUID) ([]model.PlaybackHistory, error)
	FindUserDefaultPlaylistEpisodeIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)
}

type recommendationRepository struct {
	db *gorm.DB
}

// NewRecommendationRepository は recommendationRepository を生成して RecommendationRepository として返す
func NewRecommendationRepository(db *gorm.DB) RecommendationRepository {
	return &recommendationRepository{db: db}
}

// FindRecommendedChannels は公開中チャンネルをエピソード数・再生回数・最新エピソード日で集計して取得する
func (r *recommendationRepository) FindRecommendedChannels(ctx context.Context, params RecommendChannelParams) ([]RecommendedChannel, int64, error) {
	var channels []RecommendedChannel
	var total int64

	now := time.Now()

	baseCondition := r.db.WithContext(ctx).
		Model(&model.Channel{}).
		Where("channels.published_at IS NOT NULL AND channels.published_at <= ?", now)

	if params.CategoryID != nil {
		baseCondition = baseCondition.Where("channels.category_id = ?", *params.CategoryID)
	}

	// 総件数を取得
	if err := baseCondition.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 集計データ付きで取得
	query := r.db.WithContext(ctx).
		Table("channels").
		Select(`channels.*,
			COUNT(episodes.id) as episode_count,
			COALESCE(SUM(episodes.play_count), 0) as total_play_count,
			MAX(episodes.published_at) as latest_episode_at`).
		Joins("LEFT JOIN episodes ON episodes.channel_id = channels.id AND episodes.published_at IS NOT NULL AND episodes.published_at <= ?", now).
		Where("channels.published_at IS NOT NULL AND channels.published_at <= ?", now)

	if params.CategoryID != nil {
		query = query.Where("channels.category_id = ?", *params.CategoryID)
	}

	if err := query.
		Group("channels.id").
		Order("total_play_count DESC, latest_episode_at DESC NULLS LAST").
		Limit(params.Limit).
		Offset(params.Offset).
		Find(&channels).Error; err != nil {
		return nil, 0, err
	}

	// Category と Artwork を別途取得
	if len(channels) > 0 {
		channelIDs := make([]uuid.UUID, len(channels))
		for i, ch := range channels {
			channelIDs[i] = ch.ID
		}

		// Category をプリロード
		var dbChannels []model.Channel
		if err := r.db.WithContext(ctx).
			Where("id IN ?", channelIDs).
			Preload("Category").
			Preload("Artwork").
			Find(&dbChannels).Error; err != nil {
			return nil, 0, err
		}

		channelMap := make(map[uuid.UUID]*model.Channel, len(dbChannels))
		for i := range dbChannels {
			channelMap[dbChannels[i].ID] = &dbChannels[i]
		}

		for i := range channels {
			if ch, ok := channelMap[channels[i].ID]; ok {
				channels[i].Category = ch.Category
				channels[i].Artwork = ch.Artwork
			}
		}
	}

	return channels, total, nil
}

// FindUserCategoryPreferences はユーザーの再生履歴からカテゴリ傾向を取得する
func (r *recommendationRepository) FindUserCategoryPreferences(ctx context.Context, userID uuid.UUID) ([]CategoryPreference, error) {
	var preferences []CategoryPreference

	if err := r.db.WithContext(ctx).
		Table("playback_histories").
		Select("channels.category_id, COUNT(*) as play_count").
		Joins("JOIN episodes ON episodes.id = playback_histories.episode_id").
		Joins("JOIN channels ON channels.id = episodes.channel_id").
		Where("playback_histories.user_id = ?", userID).
		Group("channels.category_id").
		Order("play_count DESC").
		Find(&preferences).Error; err != nil {
		return nil, err
	}

	return preferences, nil
}

// FindUserPlayedChannelIDs はユーザーが再生したチャンネル ID 一覧を取得する
func (r *recommendationRepository) FindUserPlayedChannelIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	var channelIDs []uuid.UUID

	if err := r.db.WithContext(ctx).
		Table("playback_histories").
		Select("DISTINCT channels.id").
		Joins("JOIN episodes ON episodes.id = playback_histories.episode_id").
		Joins("JOIN channels ON channels.id = episodes.channel_id").
		Where("playback_histories.user_id = ?", userID).
		Find(&channelIDs).Error; err != nil {
		return nil, err
	}

	return channelIDs, nil
}

// FindUserDefaultPlaylistCategoryPreferences はデフォルトプレイリスト（再生リスト）からカテゴリ傾向を取得する
func (r *recommendationRepository) FindUserDefaultPlaylistCategoryPreferences(ctx context.Context, userID uuid.UUID) ([]CategoryPreference, error) {
	var preferences []CategoryPreference

	if err := r.db.WithContext(ctx).
		Table("playlist_items").
		Select("channels.category_id, COUNT(*) as play_count").
		Joins("JOIN playlists ON playlists.id = playlist_items.playlist_id").
		Joins("JOIN episodes ON episodes.id = playlist_items.episode_id").
		Joins("JOIN channels ON channels.id = episodes.channel_id").
		Where("playlists.user_id = ? AND playlists.is_default = true", userID).
		Group("channels.category_id").
		Order("play_count DESC").
		Find(&preferences).Error; err != nil {
		return nil, err
	}

	return preferences, nil
}

// FindUserChannelIDs はユーザーが所有するチャンネル ID 一覧を取得する
func (r *recommendationRepository) FindUserChannelIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	var channelIDs []uuid.UUID

	if err := r.db.WithContext(ctx).
		Table("channels").
		Select("id").
		Where("user_id = ?", userID).
		Find(&channelIDs).Error; err != nil {
		return nil, err
	}

	return channelIDs, nil
}

// FindPublishedEpisodes は公開中エピソードを Channel・Category・Artwork・FullAudio 付きで取得する
func (r *recommendationRepository) FindPublishedEpisodes(ctx context.Context, params RecommendEpisodeParams) ([]model.Episode, int64, error) {
	var episodes []model.Episode
	var total int64

	now := time.Now()

	channelSubquery := r.db.Table("channels").Select("id").
		Where("published_at IS NOT NULL AND published_at <= ?", now)
	if params.CategoryID != nil {
		channelSubquery = channelSubquery.Where("category_id = ?", *params.CategoryID)
	}

	// 総件数を取得
	if err := r.db.WithContext(ctx).
		Model(&model.Episode{}).
		Where("published_at IS NOT NULL AND published_at <= ?", now).
		Where("channel_id IN (?)", channelSubquery).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// データを取得
	if err := r.db.WithContext(ctx).
		Where("episodes.published_at IS NOT NULL AND episodes.published_at <= ?", now).
		Where("episodes.channel_id IN (?)", channelSubquery).
		Preload("Channel").
		Preload("Channel.Category").
		Preload("Channel.Artwork").
		Preload("Artwork").
		Preload("FullAudio").
		Order("episodes.play_count DESC, episodes.published_at DESC NULLS LAST").
		Limit(params.Limit).
		Offset(params.Offset).
		Find(&episodes).Error; err != nil {
		return nil, 0, err
	}

	return episodes, total, nil
}

// FindUserPlaybackHistories はユーザーの全再生履歴を取得する
func (r *recommendationRepository) FindUserPlaybackHistories(ctx context.Context, userID uuid.UUID) ([]model.PlaybackHistory, error) {
	var histories []model.PlaybackHistory

	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("played_at DESC").
		Find(&histories).Error; err != nil {
		return nil, err
	}

	return histories, nil
}

// FindUserDefaultPlaylistEpisodeIDs はユーザーのデフォルトプレイリスト（再生リスト）のエピソード ID を取得する
func (r *recommendationRepository) FindUserDefaultPlaylistEpisodeIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	var episodeIDs []uuid.UUID

	if err := r.db.WithContext(ctx).
		Table("playlist_items").
		Select("playlist_items.episode_id").
		Joins("JOIN playlists ON playlists.id = playlist_items.playlist_id").
		Where("playlists.user_id = ? AND playlists.is_default = true", userID).
		Find(&episodeIDs).Error; err != nil {
		return nil, err
	}

	return episodeIDs, nil
}
