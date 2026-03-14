package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/pkg/cache"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
)

const (
	// episodeCacheKeyByChannel はチャンネル別エピソード一覧のキャッシュキープレフィックス。
	// 公開済み（status=published）の場合のみキャッシュする。
	episodeCacheKeyByChannel = "channel:%s:episodes:%s:%s:%d:%d"

	// episodeCacheKeyByChannelPrefix はチャンネル別エピソード一覧の無効化用プレフィックス
	episodeCacheKeyByChannelPrefix = "channel:%s:episodes:"

	episodeTTL = 5 * time.Minute
)

type cachedEpisodeRepository struct {
	repo  EpisodeRepository
	cache cache.Client
}

// NewCachedEpisodeRepository はキャッシュ付き EpisodeRepository を返す
func NewCachedEpisodeRepository(repo EpisodeRepository, cacheClient cache.Client) EpisodeRepository {
	return &cachedEpisodeRepository{repo: repo, cache: cacheClient}
}

// episodeListCache は FindByChannelID の結果をキャッシュするための構造体
type episodeListCache struct {
	Episodes []model.Episode `json:"episodes"`
	Total    int64           `json:"total"`
}

// FindByChannelID はチャンネルのエピソード一覧をキャッシュ経由で取得する。
// 公開済み（status=published）の場合のみキャッシュを使用する。
func (r *cachedEpisodeRepository) FindByChannelID(ctx context.Context, channelID uuid.UUID, filter EpisodeFilter) ([]model.Episode, int64, error) {
	if filter.Status == nil || *filter.Status != "published" {
		return r.repo.FindByChannelID(ctx, channelID, filter)
	}

	key := fmt.Sprintf(episodeCacheKeyByChannel, channelID, filter.OrderClause(), filter.Sort, filter.Limit, filter.Offset)

	var cached episodeListCache
	if hit, _ := r.cache.Get(ctx, key, &cached); hit { //nolint:errcheck // フォールバック前提
		return cached.Episodes, cached.Total, nil
	}

	episodes, total, err := r.repo.FindByChannelID(ctx, channelID, filter)
	if err != nil {
		return nil, 0, err
	}

	_ = r.cache.Set(ctx, key, episodeListCache{Episodes: episodes, Total: total}, episodeTTL) //nolint:errcheck // best effort cache
	return episodes, total, nil
}

// invalidateChannelEpisodes はチャンネルのエピソード一覧キャッシュをすべて無効化する
func (r *cachedEpisodeRepository) invalidateChannelEpisodes(ctx context.Context, channelID uuid.UUID) {
	prefix := fmt.Sprintf(episodeCacheKeyByChannelPrefix, channelID)
	_ = r.cache.DeleteByPrefix(ctx, prefix) //nolint:errcheck // best effort invalidation
}

// --- 以下は書き込み操作（キャッシュ無効化付き） ---

// Create はエピソードを作成し、チャンネルのエピソード一覧キャッシュを無効化する
func (r *cachedEpisodeRepository) Create(ctx context.Context, episode *model.Episode) error {
	err := r.repo.Create(ctx, episode)
	if err == nil {
		r.invalidateChannelEpisodes(ctx, episode.ChannelID)
	}
	return err
}

// Update はエピソードを更新し、チャンネルのエピソード一覧キャッシュを無効化する
func (r *cachedEpisodeRepository) Update(ctx context.Context, episode *model.Episode) error {
	err := r.repo.Update(ctx, episode)
	if err == nil {
		r.invalidateChannelEpisodes(ctx, episode.ChannelID)
	}
	return err
}

// Delete はエピソードを削除し、チャンネルのエピソード一覧キャッシュを無効化する。
// 削除前にエピソードを取得して channelID を特定する。
func (r *cachedEpisodeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// 削除前に channelID を取得
	episode, findErr := r.repo.FindByID(ctx, id)

	err := r.repo.Delete(ctx, id)
	if err == nil && findErr == nil {
		r.invalidateChannelEpisodes(ctx, episode.ChannelID)
	}
	return err
}

// --- 以下はキャッシュ対象外（そのままデリゲート） ---

func (r *cachedEpisodeRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Episode, error) {
	return r.repo.FindByID(ctx, id)
}

func (r *cachedEpisodeRepository) Search(ctx context.Context, filter SearchEpisodeFilter) ([]model.Episode, int64, error) {
	return r.repo.Search(ctx, filter)
}

func (r *cachedEpisodeRepository) CountPublishedByChannelIDs(ctx context.Context, channelIDs []uuid.UUID) (map[uuid.UUID]int, error) {
	return r.repo.CountPublishedByChannelIDs(ctx, channelIDs)
}

func (r *cachedEpisodeRepository) CountByChannelIDBeforeCreatedAt(ctx context.Context, channelID uuid.UUID, createdAt time.Time) (int64, error) {
	return r.repo.CountByChannelIDBeforeCreatedAt(ctx, channelID, createdAt)
}

func (r *cachedEpisodeRepository) IncrementPlayCount(ctx context.Context, id uuid.UUID) error {
	return r.repo.IncrementPlayCount(ctx, id)
}
