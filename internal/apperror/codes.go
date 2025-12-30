package apperror

import "net/http"

// エラーコードを表す型
type ErrorCode string

// エラーコード定数
const (
	CodeValidation           ErrorCode = "VALIDATION_ERROR"        // 400
	CodeReservedName         ErrorCode = "RESERVED_NAME"           // 400
	CodeScriptParse          ErrorCode = "SCRIPT_PARSE_ERROR"      // 400
	CodeSelfFollowNotAllowed ErrorCode = "SELF_FOLLOW_NOT_ALLOWED" // 400
	CodeUnauthorized         ErrorCode = "UNAUTHORIZED"            // 401
	CodeInvalidCredentials   ErrorCode = "INVALID_CREDENTIALS"     // 401
	CodeForbidden            ErrorCode = "FORBIDDEN"               // 403
	CodeNotFound             ErrorCode = "NOT_FOUND"               // 404
	CodeDuplicateEmail       ErrorCode = "DUPLICATE_EMAIL"         // 409
	CodeDuplicateUsername    ErrorCode = "DUPLICATE_USERNAME"      // 409
	CodeDuplicateName        ErrorCode = "DUPLICATE_NAME"          // 409
	CodeAlreadyLiked         ErrorCode = "ALREADY_LIKED"           // 409
	CodeAlreadyBookmarked    ErrorCode = "ALREADY_BOOKMARKED"      // 409
	CodeAlreadyFollowed      ErrorCode = "ALREADY_FOLLOWED"        // 409
	CodeSfxInUse             ErrorCode = "SFX_IN_USE"              // 409
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
	ErrValidation           = newError(CodeValidation, "Validation failed", http.StatusBadRequest)                   // バリデーションエラー
	ErrReservedName         = newError(CodeReservedName, "Reserved name cannot be used", http.StatusBadRequest)      // 予約された名前を使用
	ErrScriptParse          = newError(CodeScriptParse, "Failed to parse script", http.StatusBadRequest)             // 台本のパースに失敗
	ErrSelfFollowNotAllowed = newError(CodeSelfFollowNotAllowed, "Cannot follow own episode", http.StatusBadRequest) // 自分のエピソードはフォロー不可

	// 401 Unauthorized
	ErrUnauthorized       = newError(CodeUnauthorized, "Authentication required", http.StatusUnauthorized)         // 認証が必要
	ErrInvalidCredentials = newError(CodeInvalidCredentials, "Invalid email or password", http.StatusUnauthorized) // 認証情報が無効

	// 403 Forbidden
	ErrForbidden = newError(CodeForbidden, "Access denied", http.StatusForbidden) // アクセス権限がない

	// 404 Not Found
	ErrNotFound = newError(CodeNotFound, "Resource not found", http.StatusNotFound) // リソースが見つからない

	// 409 Conflict
	ErrDuplicateEmail    = newError(CodeDuplicateEmail, "Email already exists", http.StatusConflict)       // メールアドレスが既に使用されている
	ErrDuplicateUsername = newError(CodeDuplicateUsername, "Username already exists", http.StatusConflict) // ユーザー名が既に使用されている
	ErrDuplicateName     = newError(CodeDuplicateName, "Name already exists", http.StatusConflict)         // 名前が既に使用されている
	ErrAlreadyLiked      = newError(CodeAlreadyLiked, "Already liked", http.StatusConflict)                // 既にお気に入り済み
	ErrAlreadyBookmarked = newError(CodeAlreadyBookmarked, "Already bookmarked", http.StatusConflict)      // 既にブックマーク済み
	ErrAlreadyFollowed   = newError(CodeAlreadyFollowed, "Already followed", http.StatusConflict)          // 既にフォロー済み
	ErrSfxInUse          = newError(CodeSfxInUse, "Sound effect is in use", http.StatusConflict)           // 効果音が使用中

	// 500 Internal Server Error
	ErrInternal          = newError(CodeInternal, "Internal server error", http.StatusInternalServerError)        // サーバー内部エラー
	ErrGenerationFailed  = newError(CodeGenerationFailed, "Generation failed", http.StatusInternalServerError)    // 音声生成に失敗
	ErrMediaUploadFailed = newError(CodeMediaUploadFailed, "Media upload failed", http.StatusInternalServerError) // メディアアップロードに失敗
)
