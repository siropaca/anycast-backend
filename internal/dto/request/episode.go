package request

// 自分のチャンネルのエピソード一覧取得リクエスト
type ListMyChannelEpisodesRequest struct {
	PaginationRequest
	Status *string `form:"status" binding:"omitempty,oneof=published draft"`
}

// エピソード作成リクエスト
type CreateEpisodeRequest struct {
	Title          string  `json:"title" binding:"required,max=255"`
	Description    *string `json:"description"`
	ArtworkImageID *string `json:"artworkImageId" binding:"omitempty,uuid"`
	BgmAudioID     *string `json:"bgmAudioId" binding:"omitempty,uuid"`
}

// エピソード更新リクエスト
type UpdateEpisodeRequest struct {
	Title          *string `json:"title" binding:"omitempty,max=255"`
	Description    *string `json:"description"`
	ScriptPrompt   *string `json:"scriptPrompt"`
	ArtworkImageID *string `json:"artworkImageId" binding:"omitempty,uuid"`
	BgmAudioID     *string `json:"bgmAudioId" binding:"omitempty,uuid"`
	PublishedAt    *string `json:"publishedAt"`
}
