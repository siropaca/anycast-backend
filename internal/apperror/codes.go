package apperror

import "net/http"

// 定義済みエラー
var (
	// リソースが見つからない場合
	ErrNotFound = &AppError{
		Code:       "NOT_FOUND",
		Message:    "Resource not found",
		HTTPStatus: http.StatusNotFound,
	}
	// リクエストのバリデーションに失敗した場合
	ErrValidation = &AppError{
		Code:       "VALIDATION_ERROR",
		Message:    "Validation failed",
		HTTPStatus: http.StatusBadRequest,
	}
	// サーバー内部でエラーが発生した場合
	ErrInternal = &AppError{
		Code:       "INTERNAL_ERROR",
		Message:    "Internal server error",
		HTTPStatus: http.StatusInternalServerError,
	}
	// 認証が必要な場合
	ErrUnauthorized = &AppError{
		Code:       "UNAUTHORIZED",
		Message:    "Authentication required",
		HTTPStatus: http.StatusUnauthorized,
	}
	// アクセス権限がない場合
	ErrForbidden = &AppError{
		Code:       "FORBIDDEN",
		Message:    "Access denied",
		HTTPStatus: http.StatusForbidden,
	}
	// メールアドレスが既に使用されている場合
	ErrDuplicateEmail = &AppError{
		Code:       "DUPLICATE_EMAIL",
		Message:    "Email already exists",
		HTTPStatus: http.StatusConflict,
	}
	// ユーザー名が既に使用されている場合
	ErrDuplicateUsername = &AppError{
		Code:       "DUPLICATE_USERNAME",
		Message:    "Username already exists",
		HTTPStatus: http.StatusConflict,
	}
	// 認証情報が無効な場合
	ErrInvalidCredentials = &AppError{
		Code:       "INVALID_CREDENTIALS",
		Message:    "Invalid email or password",
		HTTPStatus: http.StatusUnauthorized,
	}
	// 名前が既に使用されている場合
	ErrDuplicateName = &AppError{
		Code:       "DUPLICATE_NAME",
		Message:    "Name already exists",
		HTTPStatus: http.StatusConflict,
	}
	// 予約された名前を使用しようとした場合
	ErrReservedName = &AppError{
		Code:       "RESERVED_NAME",
		Message:    "Reserved name cannot be used",
		HTTPStatus: http.StatusBadRequest,
	}
	// 台本のパースに失敗した場合
	ErrScriptParse = &AppError{
		Code:       "SCRIPT_PARSE_ERROR",
		Message:    "Failed to parse script",
		HTTPStatus: http.StatusBadRequest,
	}
	// 音声生成に失敗した場合
	ErrGenerationFailed = &AppError{
		Code:       "GENERATION_FAILED",
		Message:    "Generation failed",
		HTTPStatus: http.StatusInternalServerError,
	}
	// メディアのアップロードに失敗した場合
	ErrMediaUploadFailed = &AppError{
		Code:       "MEDIA_UPLOAD_FAILED",
		Message:    "Media upload failed",
		HTTPStatus: http.StatusInternalServerError,
	}
)
