package request

// 非同期台本生成リクエスト
type GenerateScriptAsyncRequest struct {
	Prompt          string `json:"prompt" binding:"max=2000"`
	DurationMinutes *int   `json:"durationMinutes" binding:"omitempty,min=3,max=30"`
	WithEmotion     bool   `json:"withEmotion"`
}

// 自分の台本生成ジョブ一覧取得リクエスト
type ListMyScriptJobsRequest struct {
	Status *string `form:"status" binding:"omitempty,oneof=pending processing completed failed"`
}

// 開発用: 台本直接生成リクエストのキャラクター情報
type GenerateScriptDirectCharacter struct {
	Name    string `json:"name" binding:"required"`
	Gender  string `json:"gender" binding:"required"`
	Persona string `json:"persona"`
}

// 開発用: DB を使わずに台本を直接生成するリクエスト
type GenerateScriptDirectRequest struct {
	// Episode
	EpisodeTitle       string `json:"episodeTitle" binding:"required"`
	EpisodeDescription string `json:"episodeDescription"`
	DurationMinutes    int    `json:"durationMinutes" binding:"required,min=3,max=30"`
	EpisodeNumber      int    `json:"episodeNumber" binding:"required,min=1"`

	// Channel
	ChannelName        string `json:"channelName" binding:"required"`
	ChannelDescription string `json:"channelDescription"`
	ChannelCategory    string `json:"channelCategory" binding:"required"`
	ChannelStyleGuide  string `json:"channelStyleGuide"`

	// Characters
	Characters []GenerateScriptDirectCharacter `json:"characters" binding:"required,min=1,dive"`

	// User
	MasterGuide string `json:"masterGuide"`

	// Theme / Options
	Theme       string `json:"theme" binding:"required,max=2000"`
	WithEmotion bool   `json:"withEmotion"`
}
