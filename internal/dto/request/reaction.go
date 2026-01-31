package request

// 高評価したエピソード一覧取得リクエスト
type ListLikesRequest struct {
	PaginationRequest
}

// リアクション登録・更新リクエスト
type CreateOrUpdateReactionRequest struct {
	ReactionType string `json:"reactionType" binding:"required,oneof=like bad"`
}
