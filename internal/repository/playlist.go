package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/pkg/logger"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
)

// PlaylistRepository は再生リストデータへのアクセスインターフェース
type PlaylistRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*model.Playlist, error)
	FindByIDWithItems(ctx context.Context, id uuid.UUID) (*model.Playlist, error)
	FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]model.Playlist, int64, error)
	ExistsByUserIDAndName(ctx context.Context, userID uuid.UUID, name string) (bool, error)
	Create(ctx context.Context, playlist *model.Playlist) error
	Update(ctx context.Context, playlist *model.Playlist) error
	Delete(ctx context.Context, id uuid.UUID) error
	CountItemsByPlaylistID(ctx context.Context, playlistID uuid.UUID) (int64, error)

	// エピソードの再生リスト所属
	FindPlaylistIDsByUserIDAndEpisodeID(ctx context.Context, userID, episodeID uuid.UUID) ([]uuid.UUID, error)

	// PlaylistItem 関連
	FindItemByID(ctx context.Context, id uuid.UUID) (*model.PlaylistItem, error)
	FindItemsByPlaylistID(ctx context.Context, playlistID uuid.UUID) ([]model.PlaylistItem, error)
	FindItemByPlaylistIDAndEpisodeID(ctx context.Context, playlistID, episodeID uuid.UUID) (*model.PlaylistItem, error)
	CreateItem(ctx context.Context, item *model.PlaylistItem) error
	DeleteItem(ctx context.Context, id uuid.UUID) error
	GetMaxPosition(ctx context.Context, playlistID uuid.UUID) (int, error)
	UpdateItemPositions(ctx context.Context, playlistID uuid.UUID, itemIDs []uuid.UUID) error
	DecrementPositionsAfter(ctx context.Context, playlistID uuid.UUID, position int) error
}

type playlistRepository struct {
	db *gorm.DB
}

// NewPlaylistRepository は PlaylistRepository の実装を返す
func NewPlaylistRepository(db *gorm.DB) PlaylistRepository {
	return &playlistRepository{db: db}
}

// FindByID は指定された ID の再生リストを取得する
func (r *playlistRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Playlist, error) {
	var playlist model.Playlist

	if err := r.db.WithContext(ctx).
		First(&playlist, "id = ?", id).Error; err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrNotFound.WithMessage("再生リストが見つかりません")
		}

		logger.FromContext(ctx).Error("failed to fetch playlist", "error", err, "playlist_id", id)
		return nil, apperror.ErrInternal.WithMessage("再生リストの取得に失敗しました").WithError(err)
	}

	return &playlist, nil
}

// FindByIDWithItems は指定された ID の再生リストをアイテムと一緒に取得する
func (r *playlistRepository) FindByIDWithItems(ctx context.Context, id uuid.UUID) (*model.Playlist, error) {
	var playlist model.Playlist

	if err := r.db.WithContext(ctx).
		Preload("Items", func(db *gorm.DB) *gorm.DB {
			return db.Order("position ASC")
		}).
		Preload("Items.Episode").
		Preload("Items.Episode.Artwork").
		Preload("Items.Episode.FullAudio").
		Preload("Items.Episode.Channel").
		Preload("Items.Episode.Channel.Artwork").
		First(&playlist, "id = ?", id).Error; err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrNotFound.WithMessage("再生リストが見つかりません")
		}

		logger.FromContext(ctx).Error("failed to fetch playlist with items", "error", err, "playlist_id", id)
		return nil, apperror.ErrInternal.WithMessage("再生リストの取得に失敗しました").WithError(err)
	}

	return &playlist, nil
}

// FindByUserID は指定されたユーザーの再生リスト一覧を取得する
func (r *playlistRepository) FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]model.Playlist, int64, error) {
	var playlists []model.Playlist
	var total int64

	tx := r.db.WithContext(ctx).Model(&model.Playlist{}).Where("user_id = ?", userID)

	// 総件数を取得
	if err := tx.Count(&total).Error; err != nil {
		logger.FromContext(ctx).Error("failed to count playlists", "error", err, "user_id", userID)
		return nil, 0, apperror.ErrInternal.WithMessage("再生リスト数の取得に失敗しました").WithError(err)
	}

	// デフォルト再生リストを先頭に、その後は作成日時順
	if err := tx.
		Order("is_default DESC, created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&playlists).Error; err != nil {
		logger.FromContext(ctx).Error("failed to fetch playlists", "error", err, "user_id", userID)
		return nil, 0, apperror.ErrInternal.WithMessage("再生リスト一覧の取得に失敗しました").WithError(err)
	}

	return playlists, total, nil
}

// ExistsByUserIDAndName は指定されたユーザーと名前の再生リストが存在するか確認する
func (r *playlistRepository) ExistsByUserIDAndName(ctx context.Context, userID uuid.UUID, name string) (bool, error) {
	var count int64

	if err := r.db.WithContext(ctx).
		Model(&model.Playlist{}).
		Where("user_id = ? AND name = ?", userID, name).
		Count(&count).Error; err != nil {
		logger.FromContext(ctx).Error("failed to check playlist existence", "error", err, "user_id", userID, "name", name)
		return false, apperror.ErrInternal.WithMessage("再生リストの確認に失敗しました").WithError(err)
	}

	return count > 0, nil
}

// Create は再生リストを作成する
func (r *playlistRepository) Create(ctx context.Context, playlist *model.Playlist) error {
	if err := r.db.WithContext(ctx).Create(playlist).Error; err != nil {
		logger.FromContext(ctx).Error("failed to create playlist", "error", err)
		return apperror.ErrInternal.WithMessage("再生リストの作成に失敗しました").WithError(err)
	}

	return nil
}

// Update は再生リストを更新する
func (r *playlistRepository) Update(ctx context.Context, playlist *model.Playlist) error {
	if err := r.db.WithContext(ctx).Save(playlist).Error; err != nil {
		logger.FromContext(ctx).Error("failed to update playlist", "error", err, "playlist_id", playlist.ID)
		return apperror.ErrInternal.WithMessage("再生リストの更新に失敗しました").WithError(err)
	}

	return nil
}

// Delete は再生リストを削除する
func (r *playlistRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&model.Playlist{}, "id = ?", id)
	if result.Error != nil {
		logger.FromContext(ctx).Error("failed to delete playlist", "error", result.Error, "playlist_id", id)
		return apperror.ErrInternal.WithMessage("再生リストの削除に失敗しました").WithError(result.Error)
	}

	if result.RowsAffected == 0 {
		return apperror.ErrNotFound.WithMessage("再生リストが見つかりません")
	}

	return nil
}

// CountItemsByPlaylistID は再生リスト内のアイテム数を取得する
func (r *playlistRepository) CountItemsByPlaylistID(ctx context.Context, playlistID uuid.UUID) (int64, error) {
	var count int64

	if err := r.db.WithContext(ctx).
		Model(&model.PlaylistItem{}).
		Where("playlist_id = ?", playlistID).
		Count(&count).Error; err != nil {
		logger.FromContext(ctx).Error("failed to count playlist items", "error", err, "playlist_id", playlistID)
		return 0, apperror.ErrInternal.WithMessage("再生リストアイテム数の取得に失敗しました").WithError(err)
	}

	return count, nil
}

// FindItemByID は指定された ID の再生リストアイテムを取得する
func (r *playlistRepository) FindItemByID(ctx context.Context, id uuid.UUID) (*model.PlaylistItem, error) {
	var item model.PlaylistItem

	if err := r.db.WithContext(ctx).
		Preload("Episode").
		Preload("Episode.Artwork").
		Preload("Episode.FullAudio").
		Preload("Episode.Channel").
		Preload("Episode.Channel.Artwork").
		First(&item, "id = ?", id).Error; err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrNotFound.WithMessage("再生リストアイテムが見つかりません")
		}

		logger.FromContext(ctx).Error("failed to fetch playlist item", "error", err, "item_id", id)
		return nil, apperror.ErrInternal.WithMessage("再生リストアイテムの取得に失敗しました").WithError(err)
	}

	return &item, nil
}

// FindItemsByPlaylistID は指定された再生リストのアイテム一覧を取得する
func (r *playlistRepository) FindItemsByPlaylistID(ctx context.Context, playlistID uuid.UUID) ([]model.PlaylistItem, error) {
	var items []model.PlaylistItem

	if err := r.db.WithContext(ctx).
		Where("playlist_id = ?", playlistID).
		Preload("Episode").
		Preload("Episode.Artwork").
		Preload("Episode.FullAudio").
		Preload("Episode.Channel").
		Preload("Episode.Channel.Artwork").
		Order("position ASC").
		Find(&items).Error; err != nil {
		logger.FromContext(ctx).Error("failed to fetch playlist items", "error", err, "playlist_id", playlistID)
		return nil, apperror.ErrInternal.WithMessage("再生リストアイテム一覧の取得に失敗しました").WithError(err)
	}

	return items, nil
}

// FindItemByPlaylistIDAndEpisodeID は指定された再生リストとエピソードのアイテムを取得する
func (r *playlistRepository) FindItemByPlaylistIDAndEpisodeID(ctx context.Context, playlistID, episodeID uuid.UUID) (*model.PlaylistItem, error) {
	var item model.PlaylistItem

	if err := r.db.WithContext(ctx).
		Where("playlist_id = ? AND episode_id = ?", playlistID, episodeID).
		First(&item).Error; err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrNotFound.WithMessage("再生リストアイテムが見つかりません")
		}

		logger.FromContext(ctx).Error("failed to fetch playlist item", "error", err, "playlist_id", playlistID, "episode_id", episodeID)
		return nil, apperror.ErrInternal.WithMessage("再生リストアイテムの取得に失敗しました").WithError(err)
	}

	return &item, nil
}

// CreateItem は再生リストアイテムを作成する
func (r *playlistRepository) CreateItem(ctx context.Context, item *model.PlaylistItem) error {
	if err := r.db.WithContext(ctx).Create(item).Error; err != nil {
		logger.FromContext(ctx).Error("failed to create playlist item", "error", err)
		return apperror.ErrInternal.WithMessage("再生リストアイテムの作成に失敗しました").WithError(err)
	}

	return nil
}

// DeleteItem は再生リストアイテムを削除する
func (r *playlistRepository) DeleteItem(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&model.PlaylistItem{}, "id = ?", id)
	if result.Error != nil {
		logger.FromContext(ctx).Error("failed to delete playlist item", "error", result.Error, "item_id", id)
		return apperror.ErrInternal.WithMessage("再生リストアイテムの削除に失敗しました").WithError(result.Error)
	}

	if result.RowsAffected == 0 {
		return apperror.ErrNotFound.WithMessage("再生リストアイテムが見つかりません")
	}

	return nil
}

// GetMaxPosition は再生リスト内の最大 position を取得する
func (r *playlistRepository) GetMaxPosition(ctx context.Context, playlistID uuid.UUID) (int, error) {
	var maxPosition *int

	if err := r.db.WithContext(ctx).
		Model(&model.PlaylistItem{}).
		Where("playlist_id = ?", playlistID).
		Select("MAX(position)").
		Scan(&maxPosition).Error; err != nil {
		logger.FromContext(ctx).Error("failed to get max position", "error", err, "playlist_id", playlistID)
		return 0, apperror.ErrInternal.WithMessage("位置情報の取得に失敗しました").WithError(err)
	}

	if maxPosition == nil {
		return -1, nil // アイテムがない場合は -1 を返す（次の position は 0 になる）
	}

	return *maxPosition, nil
}

// UpdateItemPositions は再生リストアイテムの position を一括更新する
func (r *playlistRepository) UpdateItemPositions(ctx context.Context, playlistID uuid.UUID, itemIDs []uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i, itemID := range itemIDs {
			if err := tx.Model(&model.PlaylistItem{}).
				Where("id = ? AND playlist_id = ?", itemID, playlistID).
				Update("position", i).Error; err != nil {
				logger.FromContext(ctx).Error("failed to update item position", "error", err, "item_id", itemID, "position", i)
				return apperror.ErrInternal.WithMessage("アイテムの並び替えに失敗しました").WithError(err)
			}
		}
		return nil
	})
}

// DecrementPositionsAfter は指定された position より後のアイテムの position を 1 減らす
func (r *playlistRepository) DecrementPositionsAfter(ctx context.Context, playlistID uuid.UUID, position int) error {
	if err := r.db.WithContext(ctx).
		Model(&model.PlaylistItem{}).
		Where("playlist_id = ? AND position > ?", playlistID, position).
		UpdateColumn("position", gorm.Expr("position - 1")).Error; err != nil {
		logger.FromContext(ctx).Error("failed to decrement positions", "error", err, "playlist_id", playlistID, "position", position)
		return apperror.ErrInternal.WithMessage("位置情報の更新に失敗しました").WithError(err)
	}

	return nil
}

// FindPlaylistIDsByUserIDAndEpisodeID は指定されたユーザーの再生リストのうち、エピソードが含まれる再生リスト ID 一覧を返す
func (r *playlistRepository) FindPlaylistIDsByUserIDAndEpisodeID(ctx context.Context, userID, episodeID uuid.UUID) ([]uuid.UUID, error) {
	var playlistIDs []uuid.UUID

	if err := r.db.WithContext(ctx).
		Model(&model.PlaylistItem{}).
		Select("playlist_items.playlist_id").
		Joins("JOIN playlists ON playlists.id = playlist_items.playlist_id").
		Where("playlists.user_id = ? AND playlist_items.episode_id = ?", userID, episodeID).
		Scan(&playlistIDs).Error; err != nil {
		logger.FromContext(ctx).Error("failed to fetch playlist IDs by user and episode", "error", err, "user_id", userID, "episode_id", episodeID)
		return nil, apperror.ErrInternal.WithMessage("再生リスト所属情報の取得に失敗しました").WithError(err)
	}

	return playlistIDs, nil
}
