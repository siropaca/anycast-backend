package apperror

import "fmt"

// AppError はアプリケーション全体で使用するエラー型
type AppError struct {
	Code       string      `json:"code"`
	Message    string      `json:"message"`
	HTTPStatus int         `json:"-"`
	Details    interface{} `json:"details,omitempty"`
	Err        error       `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// New は新しい AppError を作成する
func New(code string, message string, status int) *AppError {
	return &AppError{Code: code, Message: message, HTTPStatus: status}
}

// Wrap は既存のエラーをラップした AppError を作成する
func Wrap(err error, code string, message string, status int) *AppError {
	return &AppError{Code: code, Message: message, HTTPStatus: status, Err: err}
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
func (e *AppError) WithDetails(details interface{}) *AppError {
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
