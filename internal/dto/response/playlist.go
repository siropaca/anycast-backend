package response

import (
	"time"

	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
)

// プレイリスト情報のレスポンス
type PlaylistResponse struct {
	ID          uuid.UUID `json:"id" validate:"required"`
	Name        string    `json:"name" validate:"required"`
	Description string    `json:"description" validate:"required"`
	IsDefault   bool      `json:"isDefault" validate:"required"`
	ItemCount   int       `json:"itemCount" validate:"required"`
	CreatedAt   time.Time `json:"createdAt" validate:"required"`
	UpdatedAt   time.Time `json:"updatedAt" validate:"required"`
}

// プレイリスト詳細のレスポンス（アイテム含む）
type PlaylistDetailResponse struct {
	ID          uuid.UUID              `json:"id" validate:"required"`
	Name        string                 `json:"name" validate:"required"`
	Description string                 `json:"description" validate:"required"`
	IsDefault   bool                   `json:"isDefault" validate:"required"`
	Items       []PlaylistItemResponse `json:"items" validate:"required"`
	CreatedAt   time.Time              `json:"createdAt" validate:"required"`
	UpdatedAt   time.Time              `json:"updatedAt" validate:"required"`
}

// プレイリストアイテム情報のレスポンス
type PlaylistItemResponse struct {
	ID       uuid.UUID               `json:"id" validate:"required"`
	Position int                     `json:"position" validate:"required"`
	Episode  PlaylistEpisodeResponse `json:"episode" validate:"required"`
	AddedAt  time.Time               `json:"addedAt" validate:"required"`
}

// プレイリスト内のエピソード情報のレスポンス
type PlaylistEpisodeResponse struct {
	ID          uuid.UUID               `json:"id" validate:"required"`
	Title       string                  `json:"title" validate:"required"`
	Description string                  `json:"description" validate:"required"`
	Artwork     *ArtworkResponse        `json:"artwork" extensions:"x-nullable"`
	FullAudio   *AudioResponse          `json:"fullAudio" extensions:"x-nullable"`
	PlayCount   int                     `json:"playCount" validate:"required"`
	PublishedAt *time.Time              `json:"publishedAt" extensions:"x-nullable"`
	Channel     PlaylistChannelResponse `json:"channel" validate:"required"`
}

// プレイリスト内のチャンネル情報のレスポンス
type PlaylistChannelResponse struct {
	ID      uuid.UUID        `json:"id" validate:"required"`
	Name    string           `json:"name" validate:"required"`
	Artwork *ArtworkResponse `json:"artwork" extensions:"x-nullable"`
}

// プレイリスト一覧（ページネーション付き）のレスポンス
type PlaylistListWithPaginationResponse struct {
	Data       []PlaylistResponse `json:"data" validate:"required"`
	Pagination PaginationResponse `json:"pagination" validate:"required"`
}

// プレイリスト詳細のラッパーレスポンス
type PlaylistDataResponse struct {
	Data PlaylistResponse `json:"data" validate:"required"`
}

// プレイリスト詳細（アイテム含む）のラッパーレスポンス
type PlaylistDetailDataResponse struct {
	Data PlaylistDetailResponse `json:"data" validate:"required"`
}

// プレイリストアイテムのラッパーレスポンス
type PlaylistItemDataResponse struct {
	Data PlaylistItemResponse `json:"data" validate:"required"`
}
