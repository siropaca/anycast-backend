package response

// キャラクター一覧（ページネーション付き）のレスポンス
type CharacterListWithPaginationResponse struct {
	Data       []CharacterResponse `json:"data" validate:"required"`
	Pagination PaginationResponse  `json:"pagination" validate:"required"`
}

// キャラクター単体のレスポンス
type CharacterDataResponse struct {
	Data CharacterResponse `json:"data" validate:"required"`
}
