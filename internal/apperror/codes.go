package apperror

import "net/http"

// エラーコードを表す型
type ErrorCode string

// エラーコード定数
const (
	CodeValidation           ErrorCode = "VALIDATION_ERROR"        // 400
	CodeScriptParse          ErrorCode = "SCRIPT_PARSE_ERROR"      // 400
	CodeScriptTooManyLines   ErrorCode = "SCRIPT_TOO_MANY_LINES"   // 400
	CodeSelfFollowNotAllowed ErrorCode = "SELF_FOLLOW_NOT_ALLOWED" // 400
	CodeUnauthorized         ErrorCode = "UNAUTHORIZED"            // 401
	CodeInvalidCredentials   ErrorCode = "INVALID_CREDENTIALS"     // 401
	CodeInvalidRefreshToken  ErrorCode = "INVALID_REFRESH_TOKEN"   // 401
	CodeForbidden            ErrorCode = "FORBIDDEN"               // 403
	CodeNotFound             ErrorCode = "NOT_FOUND"               // 404
	CodeDuplicateEmail       ErrorCode = "DUPLICATE_EMAIL"         // 409
	CodeDuplicateUsername    ErrorCode = "DUPLICATE_USERNAME"      // 409
	CodeDuplicateName        ErrorCode = "DUPLICATE_NAME"          // 409
	CodeAlreadyLiked         ErrorCode = "ALREADY_LIKED"           // 409
	CodeAlreadyInPlaylist    ErrorCode = "ALREADY_IN_PLAYLIST"     // 409
	CodeAlreadyFollowed      ErrorCode = "ALREADY_FOLLOWED"        // 409
	CodeAlreadyFavorited     ErrorCode = "ALREADY_FAVORITED"       // 409
	CodeDefaultPlaylist      ErrorCode = "DEFAULT_PLAYLIST"        // 409
	CodeCharacterInUse       ErrorCode = "CHARACTER_IN_USE"        // 409
	CodeBgmInUse             ErrorCode = "BGM_IN_USE"              // 409
	CodeCanceled             ErrorCode = "CANCELED"                // 499
	CodeInternal             ErrorCode = "INTERNAL_ERROR"          // 500
	CodeGenerationFailed     ErrorCode = "GENERATION_FAILED"       // 500
	CodeMediaUploadFailed    ErrorCode = "MEDIA_UPLOAD_FAILED"     // 500
)

// newError は定義済みエラーを生成するヘルパー関数
func newError(code ErrorCode, message string, status int) *AppError {
	return &AppError{Code: code, Message: message, HTTPStatus: status}
}

// 定義済みエラー
var (
	// 400 Bad Request
	ErrValidation           = newError(CodeValidation, "入力内容に誤りがあります", http.StatusBadRequest)
	ErrScriptParse          = newError(CodeScriptParse, "台本の解析に失敗しました", http.StatusBadRequest)
	ErrScriptTooManyLines   = newError(CodeScriptTooManyLines, "台本の行数が上限を超えています", http.StatusBadRequest)
	ErrSelfFollowNotAllowed = newError(CodeSelfFollowNotAllowed, "自分自身はフォローできません", http.StatusBadRequest)

	// 401 Unauthorized
	ErrUnauthorized        = newError(CodeUnauthorized, "認証が必要です", http.StatusUnauthorized)
	ErrInvalidCredentials  = newError(CodeInvalidCredentials, "メールアドレスまたはパスワードが正しくありません", http.StatusUnauthorized)
	ErrInvalidRefreshToken = newError(CodeInvalidRefreshToken, "リフレッシュトークンが無効です", http.StatusUnauthorized)

	// 403 Forbidden
	ErrForbidden = newError(CodeForbidden, "アクセス権限がありません", http.StatusForbidden)

	// 404 Not Found
	ErrNotFound = newError(CodeNotFound, "リソースが見つかりません", http.StatusNotFound)

	// 409 Conflict
	ErrDuplicateEmail    = newError(CodeDuplicateEmail, "このメールアドレスは既に使用されています", http.StatusConflict)
	ErrDuplicateUsername = newError(CodeDuplicateUsername, "このユーザー名は既に使用されています", http.StatusConflict)
	ErrDuplicateName     = newError(CodeDuplicateName, "この名前は既に使用されています", http.StatusConflict)
	ErrAlreadyFollowed   = newError(CodeAlreadyFollowed, "既にフォローしています", http.StatusConflict)
	ErrAlreadyFavorited  = newError(CodeAlreadyFavorited, "既にお気に入り登録済みです", http.StatusConflict)
	ErrDefaultPlaylist   = newError(CodeDefaultPlaylist, "デフォルト再生リストは変更できません", http.StatusConflict)
	ErrCharacterInUse    = newError(CodeCharacterInUse, "このキャラクターは使用中です", http.StatusConflict)
	ErrBgmInUse          = newError(CodeBgmInUse, "この BGM は使用中です", http.StatusConflict)

	// 499 Client Closed Request（キャンセル）
	ErrCanceled = newError(CodeCanceled, "ジョブがキャンセルされました", 499)

	// 500 Internal Server Error
	ErrInternal          = newError(CodeInternal, "サーバーエラーが発生しました", http.StatusInternalServerError)
	ErrGenerationFailed  = newError(CodeGenerationFailed, "音声生成に失敗しました", http.StatusInternalServerError)
	ErrMediaUploadFailed = newError(CodeMediaUploadFailed, "メディアのアップロードに失敗しました", http.StatusInternalServerError)
)
