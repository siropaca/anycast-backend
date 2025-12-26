package apperror

import "net/http"

// 定義済みエラー
var (
	ErrNotFound = &AppError{
		Code:       "NOT_FOUND",
		Message:    "Resource not found",
		HTTPStatus: http.StatusNotFound,
	}
	ErrValidation = &AppError{
		Code:       "VALIDATION_ERROR",
		Message:    "Validation failed",
		HTTPStatus: http.StatusBadRequest,
	}
	ErrInternal = &AppError{
		Code:       "INTERNAL_ERROR",
		Message:    "Internal server error",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrUnauthorized = &AppError{
		Code:       "UNAUTHORIZED",
		Message:    "Authentication required",
		HTTPStatus: http.StatusUnauthorized,
	}
	ErrForbidden = &AppError{
		Code:       "FORBIDDEN",
		Message:    "Access denied",
		HTTPStatus: http.StatusForbidden,
	}
	ErrDuplicateEmail = &AppError{
		Code:       "DUPLICATE_EMAIL",
		Message:    "Email already exists",
		HTTPStatus: http.StatusConflict,
	}
	ErrDuplicateName = &AppError{
		Code:       "DUPLICATE_NAME",
		Message:    "Name already exists",
		HTTPStatus: http.StatusConflict,
	}
	ErrReservedName = &AppError{
		Code:       "RESERVED_NAME",
		Message:    "Reserved name cannot be used",
		HTTPStatus: http.StatusBadRequest,
	}
	ErrScriptParse = &AppError{
		Code:       "SCRIPT_PARSE_ERROR",
		Message:    "Failed to parse script",
		HTTPStatus: http.StatusBadRequest,
	}
	ErrGenerationFailed = &AppError{
		Code:       "GENERATION_FAILED",
		Message:    "Generation failed",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrMediaUploadFailed = &AppError{
		Code:       "MEDIA_UPLOAD_FAILED",
		Message:    "Media upload failed",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrSFXInUse = &AppError{
		Code:       "SFX_IN_USE",
		Message:    "Sound effect is in use",
		HTTPStatus: http.StatusConflict,
	}
)
