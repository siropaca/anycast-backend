package request

import "github.com/siropaca/anycast-backend/internal/pkg/optional"

// 自分のチャンネル一覧取得リクエスト
type ListMyChannelsRequest struct {
	PaginationRequest
	Status *string `form:"status" binding:"omitempty,oneof=published draft"`
}

// チャンネル作成リクエスト
type CreateChannelRequest struct {
	Name           string                 `json:"name" binding:"required,max=255"`
	Description    string                 `json:"description" binding:"omitempty,max=2000"`
	CategoryID     string                 `json:"categoryId" binding:"required,uuid"`
	ArtworkImageID *string                `json:"artworkImageId" binding:"omitempty,uuid"`
	Characters     ChannelCharactersInput `json:"characters" binding:"required"`
}

// チャンネルに紐づけるキャラクターの入力
type ChannelCharactersInput struct {
	Connect []ConnectCharacterInput `json:"connect" binding:"omitempty,dive"`
	Create  []CreateCharacterInput  `json:"create" binding:"omitempty,dive"`
}

// キャラクターの合計数を返す
func (c *ChannelCharactersInput) Total() int {
	return len(c.Connect) + len(c.Create)
}

// 既存キャラクターを紐づける入力
type ConnectCharacterInput struct {
	ID string `json:"id" binding:"required,uuid"`
}

// 新規キャラクター作成の入力
type CreateCharacterInput struct {
	Name     string  `json:"name" binding:"required,max=255"`
	Persona  string  `json:"persona" binding:"omitempty,max=2000"`
	AvatarID *string `json:"avatarId" binding:"omitempty,uuid"`
	VoiceID  string  `json:"voiceId" binding:"required,uuid"`
}

// チャンネル更新リクエスト
type UpdateChannelRequest struct {
	Name           string                 `json:"name" binding:"required,max=255"`
	Description    string                 `json:"description" binding:"omitempty,max=2000"`
	CategoryID     string                 `json:"categoryId" binding:"required,uuid"`
	ArtworkImageID optional.Field[string] `json:"artworkImageId"`
}

// 台本プロンプト設定リクエスト
type SetUserPromptRequest struct {
	UserPrompt string `json:"userPrompt" binding:"max=2000"`
}

// デフォルト BGM 設定リクエスト
type SetDefaultBgmRequest struct {
	BgmID       *string `json:"bgmId" binding:"omitempty,uuid"`
	SystemBgmID *string `json:"systemBgmId" binding:"omitempty,uuid"`
}

// チャンネルキャラクター追加リクエスト
// connect（既存キャラクター紐づけ）または create（新規作成）のどちらか一方を指定
type AddChannelCharacterRequest struct {
	Connect *ConnectCharacterInput `json:"connect" binding:"omitempty"`
	Create  *CreateCharacterInput  `json:"create" binding:"omitempty"`
}

// チャンネルキャラクター置換リクエスト
// connect（既存キャラクター紐づけ）または create（新規作成）のどちらか一方を指定
type ReplaceChannelCharacterRequest struct {
	Connect *ConnectCharacterInput `json:"connect" binding:"omitempty"`
	Create  *CreateCharacterInput  `json:"create" binding:"omitempty"`
}

// チャンネル公開リクエスト
type PublishChannelRequest struct {
	PublishedAt *string `json:"publishedAt"` // RFC3339 形式。省略時は現在時刻
}
