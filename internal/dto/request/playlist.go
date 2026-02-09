package request

// 再生リスト一覧取得リクエスト
type ListPlaylistsRequest struct {
	PaginationRequest
}

// 再生リスト作成リクエスト
type CreatePlaylistRequest struct {
	Name        string  `json:"name" binding:"required,max=100"`
	Description *string `json:"description" binding:"omitempty,max=500"`
}

// 再生リスト更新リクエスト
type UpdatePlaylistRequest struct {
	Name        *string `json:"name" binding:"omitempty,min=1,max=100"`
	Description *string `json:"description" binding:"omitempty,max=500"`
}

// 再生リストアイテム並び替えリクエスト
type ReorderPlaylistItemsRequest struct {
	ItemIDs []string `json:"itemIds" binding:"required,min=1,dive,uuid"`
}

// エピソードの再生リスト所属一括更新リクエスト
type UpdateEpisodePlaylistsRequest struct {
	PlaylistIDs []string `json:"playlistIds" binding:"dive,uuid"`
}
