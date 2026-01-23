package request

// 非同期音声生成リクエスト
type GenerateAudioAsyncRequest struct {
	VoiceStyle     *string  `json:"voiceStyle" binding:"omitempty,max=500"`
	BgmVolumeDB    *float64 `json:"bgmVolumeDb" binding:"omitempty,min=-60,max=0"`
	FadeOutMs      *int     `json:"fadeOutMs" binding:"omitempty,min=0,max=30000"`
	PaddingStartMs *int     `json:"paddingStartMs" binding:"omitempty,min=0,max=10000"`
	PaddingEndMs   *int     `json:"paddingEndMs" binding:"omitempty,min=0,max=10000"`
}

// 自分の音声生成ジョブ一覧取得リクエスト
type ListMyAudioJobsRequest struct {
	Status *string `form:"status" binding:"omitempty,oneof=pending processing completed failed"`
}
