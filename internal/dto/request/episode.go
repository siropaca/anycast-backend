package request

import "github.com/siropaca/anycast-backend/internal/pkg/optional"

// エピソード一覧のソートリクエスト（共通）
type EpisodeSortRequest struct {
	Sort  string `form:"sort,default=createdAt" binding:"omitempty,oneof=createdAt updatedAt"`
	Order string `form:"order,default=asc" binding:"omitempty,oneof=asc desc"`
}

// 自分のチャンネルのエピソード一覧取得リクエスト
type ListMyChannelEpisodesRequest struct {
	PaginationRequest
	EpisodeSortRequest
	Status *string `form:"status" binding:"omitempty,oneof=published draft"`
}

// チャンネルのエピソード一覧取得リクエスト
type ListChannelEpisodesRequest struct {
	PaginationRequest
	EpisodeSortRequest
}

// エピソード作成リクエスト
type CreateEpisodeRequest struct {
	Title          string  `json:"title" binding:"required,max=255"`
	Description    string  `json:"description" binding:"max=2000"`
	ArtworkImageID *string `json:"artworkImageId" binding:"omitempty,uuid"`
}

// エピソード更新リクエスト
type UpdateEpisodeRequest struct {
	Title          string                 `json:"title" binding:"required,max=255"`
	Description    string                 `json:"description" binding:"required,max=2000"`
	ArtworkImageID optional.Field[string] `json:"artworkImageId"`
}

// エピソード公開リクエスト
type PublishEpisodeRequest struct {
	PublishedAt *string `json:"publishedAt"` // RFC3339 形式。省略時は現在時刻
}

// エピソード BGM 設定リクエスト
type SetEpisodeBgmRequest struct {
	BgmID       *string `json:"bgmId" binding:"omitempty,uuid"`
	SystemBgmID *string `json:"systemBgmId" binding:"omitempty,uuid"`
}

// エピソード音声生成リクエスト
type GenerateAudioRequest struct{}
