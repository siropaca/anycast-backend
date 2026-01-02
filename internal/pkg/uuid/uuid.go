package uuid

import (
	googleuuid "github.com/google/uuid"

	"github.com/siropaca/anycast-backend/internal/apperror"
)

// UUID 型のエイリアス
type UUID = googleuuid.UUID

// Nil UUID
var Nil = googleuuid.Nil

// 文字列を UUID に変換する
// 無効な形式の場合は apperror.ErrValidation を返す
func Parse(s string) (UUID, error) {
	id, err := googleuuid.Parse(s)
	if err != nil {
		return Nil, apperror.ErrValidation.WithMessage("Invalid UUID format")
	}

	return id, nil
}

// 文字列が有効な UUID 形式かどうかを検証する
func Validate(s string) error {
	_, err := Parse(s)
	return err
}
