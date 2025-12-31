package request

// 自分のチャンネル一覧取得リクエスト
type ListMyChannelsRequest struct {
	PaginationRequest
	Status *string `form:"status" binding:"omitempty,oneof=published draft"`
}
