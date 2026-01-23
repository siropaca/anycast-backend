package apperror

import (
	"errors"
	"fmt"
)

// AppError はアプリケーション全体で使用するエラー型
type AppError struct {
	Code       ErrorCode `json:"code"`
	Message    string    `json:"message"`
	HTTPStatus int       `json:"-"`
	Details    any       `json:"details,omitempty"`
	Err        error     `json:"-"`
}

// Error は error インターフェースの実装
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap はラップされた元のエラーを返す
func (e *AppError) Unwrap() error {
	return e.Err
}

// New は新しい AppError を作成する
func New(code ErrorCode, message string, status int) *AppError {
	return &AppError{Code: code, Message: message, HTTPStatus: status}
}

// Wrap は既存のエラーをラップした AppError を作成する
func Wrap(err error, code ErrorCode, message string, status int) *AppError {
	return &AppError{Code: code, Message: message, HTTPStatus: status, Err: err}
}

// IsCode はエラーが指定したコードかどうかを判定する
func IsCode(err error, code ErrorCode) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code == code
	}
	return false
}

// WithMessage は新しいメッセージでエラーをコピーする
func (e *AppError) WithMessage(msg string) *AppError {
	return &AppError{
		Code:       e.Code,
		Message:    msg,
		HTTPStatus: e.HTTPStatus,
		Details:    e.Details,
		Err:        e.Err,
	}
}

// WithDetails は詳細情報を付与したエラーをコピーする
func (e *AppError) WithDetails(details any) *AppError {
	return &AppError{
		Code:       e.Code,
		Message:    e.Message,
		HTTPStatus: e.HTTPStatus,
		Details:    details,
		Err:        e.Err,
	}
}

// WithError は元のエラーを付与したエラーをコピーする
func (e *AppError) WithError(err error) *AppError {
	return &AppError{
		Code:       e.Code,
		Message:    e.Message,
		HTTPStatus: e.HTTPStatus,
		Details:    e.Details,
		Err:        err,
	}
}

// IsRetryable はエラーがリトライ可能かどうかを判定する
//
// 500 系のエラーはリトライ可能、400 系のエラーはリトライ不可
func IsRetryable(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.HTTPStatus >= 500
	}
	// AppError でない場合はリトライ可能と判断
	return true
}
