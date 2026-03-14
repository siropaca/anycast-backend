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
	channelCacheKeyID = "channel:%s"

	channelTTL = 10 * time.Minute
)

type cachedChannelRepository struct {
	repo  ChannelRepository
	cache cache.Client
}

// NewCachedChannelRepository はキャッシュ付き ChannelRepository を返す
func NewCachedChannelRepository(repo ChannelRepository, cacheClient cache.Client) ChannelRepository {
	return &cachedChannelRepository{repo: repo, cache: cacheClient}
}

// FindByID は指定された ID のチャンネルをキャッシュ経由で取得する
func (r *cachedChannelRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Channel, error) {
	key := fmt.Sprintf(channelCacheKeyID, id)

	var channel model.Channel
	if hit, _ := r.cache.Get(ctx, key, &channel); hit { //nolint:errcheck // フォールバック前提
		return &channel, nil
	}

	result, err := r.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	_ = r.cache.Set(ctx, key, result, channelTTL) //nolint:errcheck // best effort cache
	return result, nil
}

// invalidateChannel はチャンネルのキャッシュを無効化する
func (r *cachedChannelRepository) invalidateChannel(ctx context.Context, channelID uuid.UUID) {
	key := fmt.Sprintf(channelCacheKeyID, channelID)
	_ = r.cache.Delete(ctx, key) //nolint:errcheck // best effort invalidation
}

// --- 以下は書き込み操作（キャッシュ無効化付き） ---

// Update はチャンネルを更新し、キャッシュを無効化する
func (r *cachedChannelRepository) Update(ctx context.Context, channel *model.Channel) error {
	err := r.repo.Update(ctx, channel)
	if err == nil {
		r.invalidateChannel(ctx, channel.ID)
	}
	return err
}

// Delete はチャンネルを削除し、キャッシュを無効化する
func (r *cachedChannelRepository) Delete(ctx context.Context, id uuid.UUID) error {
	err := r.repo.Delete(ctx, id)
	if err == nil {
		r.invalidateChannel(ctx, id)
	}
	return err
}

// ReplaceChannelCharacters はチャンネルのキャラクターを置き換え、キャッシュを無効化する
func (r *cachedChannelRepository) ReplaceChannelCharacters(ctx context.Context, channelID uuid.UUID, characterIDs []uuid.UUID) error {
	err := r.repo.ReplaceChannelCharacters(ctx, channelID, characterIDs)
	if err == nil {
		r.invalidateChannel(ctx, channelID)
	}
	return err
}

// AddChannelCharacter はチャンネルにキャラクターを追加し、キャッシュを無効化する
func (r *cachedChannelRepository) AddChannelCharacter(ctx context.Context, channelID, characterID uuid.UUID) error {
	err := r.repo.AddChannelCharacter(ctx, channelID, characterID)
	if err == nil {
		r.invalidateChannel(ctx, channelID)
	}
	return err
}

// RemoveChannelCharacter はチャンネルからキャラクターを削除し、キャッシュを無効化する
func (r *cachedChannelRepository) RemoveChannelCharacter(ctx context.Context, channelID, characterID uuid.UUID) error {
	err := r.repo.RemoveChannelCharacter(ctx, channelID, characterID)
	if err == nil {
		r.invalidateChannel(ctx, channelID)
	}
	return err
}

// ReplaceChannelCharacter はチャンネル内のキャラクターを差し替え、キャッシュを無効化する
func (r *cachedChannelRepository) ReplaceChannelCharacter(ctx context.Context, channelID, oldCharacterID, newCharacterID uuid.UUID) error {
	err := r.repo.ReplaceChannelCharacter(ctx, channelID, oldCharacterID, newCharacterID)
	if err == nil {
		r.invalidateChannel(ctx, channelID)
	}
	return err
}

// --- 以下はキャッシュ対象外（そのままデリゲート） ---

func (r *cachedChannelRepository) FindByUserID(ctx context.Context, userID uuid.UUID, filter ChannelFilter) ([]model.Channel, int64, error) {
	return r.repo.FindByUserID(ctx, userID, filter)
}

func (r *cachedChannelRepository) FindPublishedByUserID(ctx context.Context, userID uuid.UUID) ([]model.Channel, error) {
	return r.repo.FindPublishedByUserID(ctx, userID)
}

func (r *cachedChannelRepository) Search(ctx context.Context, filter SearchChannelFilter) ([]model.Channel, int64, error) {
	return r.repo.Search(ctx, filter)
}

func (r *cachedChannelRepository) Create(ctx context.Context, channel *model.Channel) error {
	return r.repo.Create(ctx, channel)
}

func (r *cachedChannelRepository) CountChannelCharacters(ctx context.Context, channelID uuid.UUID) (int, error) {
	return r.repo.CountChannelCharacters(ctx, channelID)
}

func (r *cachedChannelRepository) HasChannelCharacter(ctx context.Context, channelID, characterID uuid.UUID) (bool, error) {
	return r.repo.HasChannelCharacter(ctx, channelID, characterID)
}
