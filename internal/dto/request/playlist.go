package request

// プレイリスト一覧取得リクエスト
type ListPlaylistsRequest struct {
	PaginationRequest
}

// プレイリスト作成リクエスト
type CreatePlaylistRequest struct {
	Name        string  `json:"name" binding:"required,max=100"`
	Description *string `json:"description" binding:"omitempty,max=500"`
}

// プレイリスト更新リクエスト
type UpdatePlaylistRequest struct {
	Name        *string `json:"name" binding:"omitempty,min=1,max=100"`
	Description *string `json:"description" binding:"omitempty,max=500"`
}

// プレイリストアイテム並び替えリクエスト
type ReorderPlaylistItemsRequest struct {
	ItemIDs []string `json:"itemIds" binding:"required,min=1,dive,uuid"`
}

// エピソードのプレイリスト所属一括更新リクエスト
type UpdateEpisodePlaylistsRequest struct {
	PlaylistIDs []string `json:"playlistIds" binding:"dive,uuid"`
}
