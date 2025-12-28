package apperror

import "net/http"

// エラーコードを表す型
type ErrorCode string

// エラーコード定数
const (
	CodeNotFound           ErrorCode = "NOT_FOUND"
	CodeValidation         ErrorCode = "VALIDATION_ERROR"
	CodeInternal           ErrorCode = "INTERNAL_ERROR"
	CodeUnauthorized       ErrorCode = "UNAUTHORIZED"
	CodeForbidden          ErrorCode = "FORBIDDEN"
	CodeDuplicateEmail     ErrorCode = "DUPLICATE_EMAIL"
	CodeDuplicateUsername  ErrorCode = "DUPLICATE_USERNAME"
	CodeInvalidCredentials ErrorCode = "INVALID_CREDENTIALS"
	CodeDuplicateName      ErrorCode = "DUPLICATE_NAME"
	CodeReservedName       ErrorCode = "RESERVED_NAME"
	CodeScriptParse        ErrorCode = "SCRIPT_PARSE_ERROR"
	CodeGenerationFailed   ErrorCode = "GENERATION_FAILED"
	CodeMediaUploadFailed  ErrorCode = "MEDIA_UPLOAD_FAILED"
)

// 定義済みエラー
var (
	// リソースが見つからない場合
	ErrNotFound = &AppError{
		Code:       CodeNotFound,
		Message:    "Resource not found",
		HTTPStatus: http.StatusNotFound,
	}
	// リクエストのバリデーションに失敗した場合
	ErrValidation = &AppError{
		Code:       CodeValidation,
		Message:    "Validation failed",
		HTTPStatus: http.StatusBadRequest,
	}
	// サーバー内部でエラーが発生した場合
	ErrInternal = &AppError{
		Code:       CodeInternal,
		Message:    "Internal server error",
		HTTPStatus: http.StatusInternalServerError,
	}
	// 認証が必要な場合
	ErrUnauthorized = &AppError{
		Code:       CodeUnauthorized,
		Message:    "Authentication required",
		HTTPStatus: http.StatusUnauthorized,
	}
	// アクセス権限がない場合
	ErrForbidden = &AppError{
		Code:       CodeForbidden,
		Message:    "Access denied",
		HTTPStatus: http.StatusForbidden,
	}
	// メールアドレスが既に使用されている場合
	ErrDuplicateEmail = &AppError{
		Code:       CodeDuplicateEmail,
		Message:    "Email already exists",
		HTTPStatus: http.StatusConflict,
	}
	// ユーザー名が既に使用されている場合
	ErrDuplicateUsername = &AppError{
		Code:       CodeDuplicateUsername,
		Message:    "Username already exists",
		HTTPStatus: http.StatusConflict,
	}
	// 認証情報が無効な場合
	ErrInvalidCredentials = &AppError{
		Code:       CodeInvalidCredentials,
		Message:    "Invalid email or password",
		HTTPStatus: http.StatusUnauthorized,
	}
	// 名前が既に使用されている場合
	ErrDuplicateName = &AppError{
		Code:       CodeDuplicateName,
		Message:    "Name already exists",
		HTTPStatus: http.StatusConflict,
	}
	// 予約された名前を使用しようとした場合
	ErrReservedName = &AppError{
		Code:       CodeReservedName,
		Message:    "Reserved name cannot be used",
		HTTPStatus: http.StatusBadRequest,
	}
	// 台本のパースに失敗した場合
	ErrScriptParse = &AppError{
		Code:       CodeScriptParse,
		Message:    "Failed to parse script",
		HTTPStatus: http.StatusBadRequest,
	}
	// 音声生成に失敗した場合
	ErrGenerationFailed = &AppError{
		Code:       CodeGenerationFailed,
		Message:    "Generation failed",
		HTTPStatus: http.StatusInternalServerError,
	}
	// メディアのアップロードに失敗した場合
	ErrMediaUploadFailed = &AppError{
		Code:       CodeMediaUploadFailed,
		Message:    "Media upload failed",
		HTTPStatus: http.StatusInternalServerError,
	}
)
