package request

// 自分のチャンネルのエピソード一覧取得リクエスト
type ListMyChannelEpisodesRequest struct {
	PaginationRequest
	Status *string `form:"status" binding:"omitempty,oneof=published draft"`
}

// チャンネルのエピソード一覧取得リクエスト
type ListChannelEpisodesRequest struct {
	PaginationRequest
}

// エピソード作成リクエスト
type CreateEpisodeRequest struct {
	Title          string  `json:"title" binding:"required,max=255"`
	Description    string  `json:"description" binding:"max=2000"`
	ArtworkImageID *string `json:"artworkImageId" binding:"omitempty,uuid"`
}

// エピソード更新リクエスト
type UpdateEpisodeRequest struct {
	Title          string  `json:"title" binding:"required,max=255"`
	Description    string  `json:"description" binding:"required,max=2000"`
	ArtworkImageID *string `json:"artworkImageId" binding:"omitempty,uuid"`
}

// エピソード公開リクエスト
type PublishEpisodeRequest struct {
	PublishedAt *string `json:"publishedAt"` // RFC3339 形式。省略時は現在時刻
}

// エピソード音声生成リクエスト
type GenerateAudioRequest struct {
	VoiceStyle *string `json:"voiceStyle" binding:"omitempty,max=500"`
}
