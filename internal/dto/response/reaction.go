package response

import (
	"time"

	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
)

// 高評価一覧内のエピソード情報のレスポンス
type LikeEpisodeResponse struct {
	ID          uuid.UUID           `json:"id" validate:"required"`
	Title       string              `json:"title" validate:"required"`
	Description string              `json:"description" validate:"required"`
	Channel     LikeChannelResponse `json:"channel" validate:"required"`
	PublishedAt *time.Time          `json:"publishedAt" extensions:"x-nullable"`
}

// 高評価一覧内のチャンネル情報のレスポンス
type LikeChannelResponse struct {
	ID      uuid.UUID        `json:"id" validate:"required"`
	Name    string           `json:"name" validate:"required"`
	Artwork *ArtworkResponse `json:"artwork" extensions:"x-nullable"`
}

// 高評価アイテムのレスポンス
type LikeItemResponse struct {
	Episode LikeEpisodeResponse `json:"episode" validate:"required"`
	LikedAt time.Time           `json:"likedAt" validate:"required"`
}

// 高評価一覧（ページネーション付き）のレスポンス
type LikeListWithPaginationResponse struct {
	Data       []LikeItemResponse `json:"data" validate:"required"`
	Pagination PaginationResponse `json:"pagination" validate:"required"`
}

// リアクションのレスポンス
type ReactionResponse struct {
	ID           uuid.UUID `json:"id" validate:"required"`
	EpisodeID    uuid.UUID `json:"episodeId" validate:"required"`
	ReactionType string    `json:"reactionType" validate:"required"`
	CreatedAt    time.Time `json:"createdAt" validate:"required"`
}

// リアクション登録・更新のラッパーレスポンス
type ReactionDataResponse struct {
	Data ReactionResponse `json:"data" validate:"required"`
}

// リアクション状態のレスポンス
type ReactionStatusResponse struct {
	ReactionType *string `json:"reactionType" extensions:"x-nullable"`
}

// リアクション状態のラッパーレスポンス
type ReactionStatusDataResponse struct {
	Data ReactionStatusResponse `json:"data" validate:"required"`
}
