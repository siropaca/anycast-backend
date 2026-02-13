package request

// AI 画像生成リクエスト
type GenerateImageRequest struct {
	Prompt string `json:"prompt" binding:"required,max=1000"`
}
