package request

// 台本生成リクエスト
type GenerateScriptRequest struct {
	Prompt          string `json:"prompt" binding:"required,max=2000"` // Episode の UserPrompt になる
	DurationMinutes *int   `json:"durationMinutes" binding:"omitempty,min=3,max=30"`
	WithEmotion     bool   `json:"withEmotion"` // 感情を付与するかどうか
}

// 台本インポートリクエスト
type ImportScriptRequest struct {
	Text string `json:"text" binding:"required"`
}
