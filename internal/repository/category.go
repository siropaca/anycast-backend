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

// CategoryStats はカテゴリごとの公開済みチャンネル数・エピソード数を保持する
type CategoryStats struct {
	CategoryID   uuid.UUID
	ChannelCount int
	EpisodeCount int
}

// CategoryRepository はカテゴリデータへのアクセスインターフェース
type CategoryRepository interface {
	FindAllActive(ctx context.Context) ([]model.Category, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.Category, error)
	FindBySlug(ctx context.Context, slug string) (*model.Category, error)
	FindStatsByCategoryIDs(ctx context.Context, categoryIDs []uuid.UUID) (map[uuid.UUID]CategoryStats, error)
}

type categoryRepository struct {
	db *gorm.DB
}

// NewCategoryRepository は CategoryRepository の実装を返す
func NewCategoryRepository(db *gorm.DB) CategoryRepository {
	return &categoryRepository{db: db}
}

// FindAllActive はアクティブなカテゴリを表示順で取得する
func (r *categoryRepository) FindAllActive(ctx context.Context) ([]model.Category, error) {
	var categories []model.Category

	if err := r.db.WithContext(ctx).Preload("Image").Where("is_active = ?", true).Order("sort_order ASC").Find(&categories).Error; err != nil {
		logger.FromContext(ctx).Error("failed to fetch categories", "error", err)
		return nil, apperror.ErrInternal.WithMessage("カテゴリ一覧の取得に失敗しました").WithError(err)
	}

	return categories, nil
}

// FindByID は指定された ID のカテゴリを取得する
func (r *categoryRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Category, error) {
	var category model.Category

	if err := r.db.WithContext(ctx).Preload("Image").First(&category, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrNotFound.WithMessage("カテゴリが見つかりません")
		}
		logger.FromContext(ctx).Error("failed to fetch category", "error", err, "category_id", id)
		return nil, apperror.ErrInternal.WithMessage("カテゴリの取得に失敗しました").WithError(err)
	}

	return &category, nil
}

// FindBySlug は指定されたスラッグのカテゴリを取得する
func (r *categoryRepository) FindBySlug(ctx context.Context, slug string) (*model.Category, error) {
	var category model.Category

	if err := r.db.WithContext(ctx).Preload("Image").First(&category, "slug = ?", slug).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrNotFound.WithMessage("カテゴリが見つかりません")
		}
		logger.FromContext(ctx).Error("failed to fetch category by slug", "error", err, "slug", slug)
		return nil, apperror.ErrInternal.WithMessage("カテゴリの取得に失敗しました").WithError(err)
	}

	return &category, nil
}

// FindStatsByCategoryIDs は指定されたカテゴリ群ごとの公開済みチャンネル数・エピソード数を一括取得する
func (r *categoryRepository) FindStatsByCategoryIDs(ctx context.Context, categoryIDs []uuid.UUID) (map[uuid.UUID]CategoryStats, error) {
	if len(categoryIDs) == 0 {
		return map[uuid.UUID]CategoryStats{}, nil
	}

	var rows []CategoryStats
	now := time.Now()

	if err := r.db.WithContext(ctx).
		Table("channels").
		Select("channels.category_id, COUNT(DISTINCT channels.id) AS channel_count, COUNT(episodes.id) AS episode_count").
		Joins("LEFT JOIN episodes ON episodes.channel_id = channels.id AND episodes.published_at IS NOT NULL AND episodes.published_at <= ?", now).
		Where("channels.category_id IN ? AND channels.published_at IS NOT NULL AND channels.published_at <= ?", categoryIDs, now).
		Group("channels.category_id").
		Find(&rows).Error; err != nil {
		logger.FromContext(ctx).Error("failed to fetch category stats", "error", err)
		return nil, apperror.ErrInternal.WithMessage("カテゴリ統計情報の取得に失敗しました").WithError(err)
	}

	result := make(map[uuid.UUID]CategoryStats, len(rows))
	for _, row := range rows {
		result[row.CategoryID] = row
	}

	return result, nil
}
