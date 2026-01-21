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

// 台本行更新リクエスト
type UpdateScriptLineRequest struct {
	SpeakerID *string `json:"speakerId" binding:"omitempty,uuid"`
	Text      *string `json:"text"`
	Emotion   *string `json:"emotion"`
}

// 台本行作成リクエスト
type CreateScriptLineRequest struct {
	SpeakerID   string  `json:"speakerId" binding:"required,uuid"`
	Text        string  `json:"text"`
	Emotion     *string `json:"emotion"`
	AfterLineID *string `json:"afterLineId" binding:"omitempty,uuid"`
}

// 台本行並び替えリクエスト
type ReorderScriptLinesRequest struct {
	LineIDs []string `json:"lineIds" binding:"required,min=1,dive,uuid"`
}
