package request

// 台本インポートリクエスト
type ImportScriptRequest struct {
	Text string `json:"text" binding:"required"`
}

// 台本行更新リクエスト
type UpdateScriptLineRequest struct {
	SpeakerID *string `json:"speakerId" binding:"omitempty,uuid"`
	Text      *string `json:"text" binding:"omitempty,max=500"`
	Emotion   *string `json:"emotion" binding:"omitempty,max=20"`
}

// 台本行作成リクエスト
type CreateScriptLineRequest struct {
	SpeakerID   string  `json:"speakerId" binding:"required,uuid"`
	Text        string  `json:"text" binding:"max=500"`
	Emotion     *string `json:"emotion" binding:"omitempty,max=20"`
	AfterLineID *string `json:"afterLineId" binding:"omitempty,uuid"`
}

// 台本行並び替えリクエスト
type ReorderScriptLinesRequest struct {
	LineIDs []string `json:"lineIds" binding:"required,min=1,dive,uuid"`
}
