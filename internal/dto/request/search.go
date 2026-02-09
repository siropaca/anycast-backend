package request

// チャンネル検索リクエスト
type SearchChannelsRequest struct {
	PaginationRequest
	Q            string  `form:"q" binding:"required,max=255"`
	CategorySlug *string `form:"categorySlug" binding:"omitempty,max=50"`
}

// エピソード検索リクエスト
type SearchEpisodesRequest struct {
	PaginationRequest
	Q string `form:"q" binding:"required,max=255"`
}

// ユーザー検索リクエスト
type SearchUsersRequest struct {
	PaginationRequest
	Q string `form:"q" binding:"required,max=255"`
}
