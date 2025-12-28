package apperror

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

const testCode ErrorCode = "TEST_CODE"

func TestNew(t *testing.T) {
	t.Run("指定した値で AppError を作成できる", func(t *testing.T) {
		err := New(testCode, "test message", 400)

		assert.Equal(t, testCode, err.Code)
		assert.Equal(t, "test message", err.Message)
		assert.Equal(t, 400, err.HTTPStatus)
		assert.Nil(t, err.Details)
		assert.Nil(t, err.Err)
	})
}

func TestWrap(t *testing.T) {
	t.Run("既存のエラーをラップした AppError を作成できる", func(t *testing.T) {
		innerErr := errors.New("inner error")
		err := Wrap(innerErr, testCode, "wrap message", 500)

		assert.Equal(t, testCode, err.Code)
		assert.Equal(t, "wrap message", err.Message)
		assert.Equal(t, 500, err.HTTPStatus)
		assert.Equal(t, innerErr, err.Err)
	})
}

func TestIsCode(t *testing.T) {
	t.Run("AppError のコードが一致する場合は true を返す", func(t *testing.T) {
		err := New(CodeNotFound, "not found", 404)

		assert.True(t, IsCode(err, CodeNotFound))
	})

	t.Run("AppError のコードが一致しない場合は false を返す", func(t *testing.T) {
		err := New(CodeNotFound, "not found", 404)

		assert.False(t, IsCode(err, CodeInternal))
	})

	t.Run("AppError でない場合は false を返す", func(t *testing.T) {
		err := errors.New("standard error")

		assert.False(t, IsCode(err, CodeNotFound))
	})

	t.Run("nil の場合は false を返す", func(t *testing.T) {
		assert.False(t, IsCode(nil, CodeNotFound))
	})
}

func TestAppError_Error(t *testing.T) {
	t.Run("ラップされたエラーがない場合はコードとメッセージを返す", func(t *testing.T) {
		err := New(testCode, "message", 400)

		assert.Equal(t, "TEST_CODE: message", err.Error())
	})

	t.Run("ラップされたエラーがある場合は含めて返す", func(t *testing.T) {
		innerErr := errors.New("inner")
		err := Wrap(innerErr, testCode, "message", 500)

		assert.Equal(t, "TEST_CODE: message: inner", err.Error())
	})
}

func TestAppError_Unwrap(t *testing.T) {
	t.Run("ラップされたエラーを返す", func(t *testing.T) {
		innerErr := errors.New("inner")
		err := Wrap(innerErr, testCode, "message", 500)

		assert.Equal(t, innerErr, err.Unwrap())
	})

	t.Run("ラップされたエラーがない場合は nil を返す", func(t *testing.T) {
		err := New(testCode, "message", 400)

		assert.Nil(t, err.Unwrap())
	})

	t.Run("errors.Is で判定できる", func(t *testing.T) {
		innerErr := errors.New("inner")
		err := Wrap(innerErr, testCode, "message", 500)

		assert.ErrorIs(t, err, innerErr)
	})
}

func TestAppError_WithMessage(t *testing.T) {
	t.Run("新しいメッセージを持つコピーを返す", func(t *testing.T) {
		original := New(testCode, "old message", 400)
		updated := original.WithMessage("new message")

		assert.Equal(t, "new message", updated.Message)
		assert.Equal(t, "old message", original.Message, "オリジナルは変更されない")
	})

	t.Run("他のフィールドは保持される", func(t *testing.T) {
		innerErr := errors.New("inner")
		original := Wrap(innerErr, testCode, "message", 500)
		original.Details = "details"

		updated := original.WithMessage("new")

		assert.Equal(t, original.Code, updated.Code)
		assert.Equal(t, original.HTTPStatus, updated.HTTPStatus)
		assert.Equal(t, original.Details, updated.Details)
		assert.Equal(t, original.Err, updated.Err)
	})
}

func TestAppError_WithDetails(t *testing.T) {
	t.Run("詳細情報を持つコピーを返す", func(t *testing.T) {
		original := New(testCode, "message", 400)
		details := map[string]string{"field": "value"}
		updated := original.WithDetails(details)

		assert.Equal(t, details, updated.Details)
		assert.Nil(t, original.Details, "オリジナルは変更されない")
	})

	t.Run("他のフィールドは保持される", func(t *testing.T) {
		original := New(testCode, "message", 400)
		updated := original.WithDetails("details")

		assert.Equal(t, original.Code, updated.Code)
		assert.Equal(t, original.Message, updated.Message)
		assert.Equal(t, original.HTTPStatus, updated.HTTPStatus)
	})
}

func TestAppError_WithError(t *testing.T) {
	t.Run("元のエラーを持つコピーを返す", func(t *testing.T) {
		original := New(testCode, "message", 400)
		innerErr := errors.New("inner")
		updated := original.WithError(innerErr)

		assert.Equal(t, innerErr, updated.Err)
		assert.Nil(t, original.Err, "オリジナルは変更されない")
	})

	t.Run("他のフィールドは保持される", func(t *testing.T) {
		original := New(testCode, "message", 400)
		original.Details = "details"
		innerErr := errors.New("inner")
		updated := original.WithError(innerErr)

		assert.Equal(t, original.Code, updated.Code)
		assert.Equal(t, original.Message, updated.Message)
		assert.Equal(t, original.HTTPStatus, updated.HTTPStatus)
		assert.Equal(t, original.Details, updated.Details)
	})
}
