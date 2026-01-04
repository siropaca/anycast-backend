package request

// 自分のチャンネル一覧取得リクエスト
type ListMyChannelsRequest struct {
	PaginationRequest
	Status *string `form:"status" binding:"omitempty,oneof=published draft"`
}

// チャンネル作成リクエスト
type CreateChannelRequest struct {
	Name           string                   `json:"name" binding:"required,max=255"`
	Description    string                   `json:"description" binding:"required,max=2000"`
	UserPrompt     string                   `json:"userPrompt" binding:"required,max=2000"`
	CategoryID     string                   `json:"categoryId" binding:"required,uuid"`
	ArtworkImageID *string                  `json:"artworkImageId" binding:"omitempty,uuid"`
	Characters     []CreateCharacterRequest `json:"characters" binding:"required,min=1,max=2,dive"`
}

// キャラクター作成リクエスト
type CreateCharacterRequest struct {
	Name    string `json:"name" binding:"required,max=255"`
	Persona string `json:"persona" binding:"max=2000"`
	VoiceID string `json:"voiceId" binding:"required,uuid"`
}

// チャンネル更新リクエスト
type UpdateChannelRequest struct {
	Name           *string `json:"name" binding:"omitempty,max=255"`
	Description    *string `json:"description" binding:"omitempty,max=2000"`
	UserPrompt     *string `json:"userPrompt" binding:"omitempty,max=2000"`
	CategoryID     *string `json:"categoryId" binding:"omitempty,uuid"`
	ArtworkImageID *string `json:"artworkImageId" binding:"omitempty,uuid"`
}

// チャンネル公開リクエスト
type PublishChannelRequest struct {
	PublishedAt *string `json:"publishedAt"` // RFC3339 形式。省略時は現在時刻
}
