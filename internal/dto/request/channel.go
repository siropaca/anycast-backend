package request

// 自分のチャンネル一覧取得リクエスト
type ListMyChannelsRequest struct {
	PaginationRequest
	Status *string `form:"status" binding:"omitempty,oneof=published draft"`
}

// チャンネル作成リクエスト
type CreateChannelRequest struct {
	Name           string                   `json:"name" binding:"required,max=255"`
	Description    string                   `json:"description" binding:"required"`
	ScriptPrompt   string                   `json:"scriptPrompt" binding:"required"`
	CategoryID     string                   `json:"categoryId" binding:"required,uuid"`
	ArtworkImageID *string                  `json:"artworkImageId" binding:"omitempty,uuid"`
	Characters     []CreateCharacterRequest `json:"characters" binding:"required,min=1,max=2,dive"`
}

// キャラクター作成リクエスト
type CreateCharacterRequest struct {
	Name    string `json:"name" binding:"required,max=255"`
	Persona string `json:"persona"`
	VoiceID string `json:"voiceId" binding:"required,uuid"`
}
