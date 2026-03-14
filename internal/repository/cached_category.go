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
	categoryCacheKeyActive = "categories:active"
	categoryCacheKeyStats  = "categories:stats"
	categoryCacheKeySlug   = "category:slug:%s"

	categoryTTL = 1 * time.Hour
)

type cachedCategoryRepository struct {
	repo  CategoryRepository
	cache cache.Client
}

// NewCachedCategoryRepository はキャッシュ付き CategoryRepository を返す
func NewCachedCategoryRepository(repo CategoryRepository, cacheClient cache.Client) CategoryRepository {
	return &cachedCategoryRepository{repo: repo, cache: cacheClient}
}

// FindAllActive はアクティブなカテゴリ一覧をキャッシュ経由で取得する
func (r *cachedCategoryRepository) FindAllActive(ctx context.Context) ([]model.Category, error) {
	var categories []model.Category
	if hit, _ := r.cache.Get(ctx, categoryCacheKeyActive, &categories); hit { //nolint:errcheck // フォールバック前提
		return categories, nil
	}

	categories, err := r.repo.FindAllActive(ctx)
	if err != nil {
		return nil, err
	}

	_ = r.cache.Set(ctx, categoryCacheKeyActive, categories, categoryTTL) //nolint:errcheck // best effort cache
	return categories, nil
}

// FindByID は指定された ID のカテゴリを取得する（キャッシュ対象外）
func (r *cachedCategoryRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Category, error) {
	return r.repo.FindByID(ctx, id)
}

// FindBySlug は指定されたスラッグのカテゴリをキャッシュ経由で取得する
func (r *cachedCategoryRepository) FindBySlug(ctx context.Context, slug string) (*model.Category, error) {
	key := fmt.Sprintf(categoryCacheKeySlug, slug)

	var category model.Category
	if hit, _ := r.cache.Get(ctx, key, &category); hit { //nolint:errcheck // フォールバック前提
		return &category, nil
	}

	result, err := r.repo.FindBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	_ = r.cache.Set(ctx, key, result, categoryTTL) //nolint:errcheck // best effort cache
	return result, nil
}

// FindStatsByCategoryIDs はカテゴリ統計をキャッシュ経由で取得する
func (r *cachedCategoryRepository) FindStatsByCategoryIDs(ctx context.Context, categoryIDs []uuid.UUID) (map[uuid.UUID]CategoryStats, error) {
	var statsMap map[uuid.UUID]CategoryStats
	if hit, _ := r.cache.Get(ctx, categoryCacheKeyStats, &statsMap); hit { //nolint:errcheck // フォールバック前提
		return statsMap, nil
	}

	statsMap, err := r.repo.FindStatsByCategoryIDs(ctx, categoryIDs)
	if err != nil {
		return nil, err
	}

	_ = r.cache.Set(ctx, categoryCacheKeyStats, statsMap, categoryTTL) //nolint:errcheck // best effort cache
	return statsMap, nil
}
