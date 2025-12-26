package apperror

import "fmt"

// アプリケーション全体で使用するエラー型
type AppError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	HTTPStatus int    `json:"-"`
	Details    any    `json:"details,omitempty"`
	Err        error  `json:"-"`
}

// error インターフェースの実装
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// ラップされた元のエラーを返す
func (e *AppError) Unwrap() error {
	return e.Err
}

// 新しい AppError を作成する
func New(code string, message string, status int) *AppError {
	return &AppError{Code: code, Message: message, HTTPStatus: status}
}

// 既存のエラーをラップした AppError を作成する
func Wrap(err error, code string, message string, status int) *AppError {
	return &AppError{Code: code, Message: message, HTTPStatus: status, Err: err}
}

// 新しいメッセージでエラーをコピーする
func (e *AppError) WithMessage(msg string) *AppError {
	return &AppError{
		Code:       e.Code,
		Message:    msg,
		HTTPStatus: e.HTTPStatus,
		Details:    e.Details,
		Err:        e.Err,
	}
}

// 詳細情報を付与したエラーをコピーする
func (e *AppError) WithDetails(details any) *AppError {
	return &AppError{
		Code:       e.Code,
		Message:    e.Message,
		HTTPStatus: e.HTTPStatus,
		Details:    details,
		Err:        e.Err,
	}
}

// 元のエラーを付与したエラーをコピーする
func (e *AppError) WithError(err error) *AppError {
	return &AppError{
		Code:       e.Code,
		Message:    e.Message,
		HTTPStatus: e.HTTPStatus,
		Details:    e.Details,
		Err:        err,
	}
}
