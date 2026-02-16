package request

import "github.com/siropaca/anycast-backend/internal/pkg/optional"

// 自分のキャラクター一覧取得リクエスト
type ListMyCharactersRequest struct {
	PaginationRequest
}

// キャラクター作成リクエスト
type CreateCharacterRequest struct {
	Name     string  `json:"name" binding:"required,max=255"`
	Persona  string  `json:"persona" binding:"max=2000"`
	AvatarID *string `json:"avatarId" binding:"omitempty,uuid"`
	VoiceID  string  `json:"voiceId" binding:"required,uuid"`
}

// キャラクター更新リクエスト
type UpdateCharacterRequest struct {
	Name     *string                `json:"name" binding:"omitempty,max=255"`
	Persona  *string                `json:"persona" binding:"omitempty,max=2000"`
	AvatarID optional.Field[string] `json:"avatarId"`
	VoiceID  *string                `json:"voiceId" binding:"omitempty,uuid"`
}
