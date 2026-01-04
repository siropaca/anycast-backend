package request

// 台本生成リクエスト
type GenerateScriptRequest struct {
	Prompt          string `json:"prompt" binding:"required,max=2000"`
	DurationMinutes *int   `json:"durationMinutes" binding:"omitempty,min=3,max=30"`
}
