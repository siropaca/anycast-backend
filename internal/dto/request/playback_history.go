package request

// 再生履歴一覧取得リクエスト
type ListPlaybackHistoryRequest struct {
	PaginationRequest
	Completed *bool `form:"completed"`
}

// 再生履歴更新リクエスト
type UpdatePlaybackRequest struct {
	ProgressMs *int  `json:"progressMs" binding:"omitempty,min=0"`
	Completed  *bool `json:"completed"`
}
