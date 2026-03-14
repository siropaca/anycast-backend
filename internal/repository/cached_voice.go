package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/pkg/cache"
)

const (
	voiceCacheKeyAll = "voices:all"
	voiceCacheKeyID  = "voice:%s"

	voiceTTL = 2 * time.Hour
)

type cachedVoiceRepository struct {
	repo  VoiceRepository
	cache cache.Client
}

// NewCachedVoiceRepository はキャッシュ付き VoiceRepository を返す
func NewCachedVoiceRepository(repo VoiceRepository, cacheClient cache.Client) VoiceRepository {
	return &cachedVoiceRepository{repo: repo, cache: cacheClient}
}

// FindAll はフィルタ条件に基づいてボイス一覧をキャッシュ経由で取得する。
// フィルタなしの場合のみキャッシュを使用する。
func (r *cachedVoiceRepository) FindAll(ctx context.Context, filter VoiceFilter) ([]model.Voice, error) {
	if filter.Provider != nil || filter.Gender != nil {
		return r.repo.FindAll(ctx, filter)
	}

	var voices []model.Voice
	if hit, _ := r.cache.Get(ctx, voiceCacheKeyAll, &voices); hit { //nolint:errcheck // フォールバック前提
		return voices, nil
	}

	voices, err := r.repo.FindAll(ctx, filter)
	if err != nil {
		return nil, err
	}

	_ = r.cache.Set(ctx, voiceCacheKeyAll, voices, voiceTTL) //nolint:errcheck // best effort cache
	return voices, nil
}

// FindByID は指定された ID のボイスをキャッシュ経由で取得する
func (r *cachedVoiceRepository) FindByID(ctx context.Context, id string) (*model.Voice, error) {
	key := fmt.Sprintf(voiceCacheKeyID, id)

	var voice model.Voice
	if hit, _ := r.cache.Get(ctx, key, &voice); hit { //nolint:errcheck // フォールバック前提
		return &voice, nil
	}

	result, err := r.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	_ = r.cache.Set(ctx, key, result, voiceTTL) //nolint:errcheck // best effort cache
	return result, nil
}

// FindActiveByID は指定された ID のアクティブなボイスを取得する（キャッシュ対象外）
func (r *cachedVoiceRepository) FindActiveByID(ctx context.Context, id string) (*model.Voice, error) {
	return r.repo.FindActiveByID(ctx, id)
}
