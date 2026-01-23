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
