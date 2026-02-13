package request

// チャンネルアートワーク生成リクエスト
type GenerateChannelArtworkRequest struct {
	Prompt     *string `json:"prompt" binding:"omitempty,max=1000"`
	SetArtwork *bool   `json:"setArtwork"`
}

// エピソードアートワーク生成リクエスト
type GenerateEpisodeArtworkRequest struct {
	Prompt     *string `json:"prompt" binding:"omitempty,max=1000"`
	SetArtwork *bool   `json:"setArtwork"`
}
