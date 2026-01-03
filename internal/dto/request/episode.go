package request

// 自分のチャンネルのエピソード一覧取得リクエスト
type ListMyChannelEpisodesRequest struct {
	PaginationRequest
	Status *string `form:"status" binding:"omitempty,oneof=published draft"`
}
