package request

// 自分の BGM 一覧取得リクエスト
type ListMyBgmsRequest struct {
	PaginationRequest
	IncludeSystem bool `form:"include_system,default=false"`
}

// BGM 作成リクエスト
type CreateBgmRequest struct {
	Name    string `json:"name" binding:"required,max=255"`
	AudioID string `json:"audioId" binding:"required,uuid"`
}

// BGM 更新リクエスト
type UpdateBgmRequest struct {
	Name *string `json:"name" binding:"omitempty,max=255"`
}
