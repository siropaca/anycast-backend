package uuid

import (
	googleuuid "github.com/google/uuid"

	"github.com/siropaca/anycast-backend/internal/apperror"
)

// UUID は UUID 型のエイリアス
type UUID = googleuuid.UUID

// Nil UUID
var Nil = googleuuid.Nil

// New は新しい UUID を生成する
func New() UUID {
	return googleuuid.New()
}

// Parse は文字列を UUID に変換する
// 無効な形式の場合は apperror.ErrValidation を返す
func Parse(s string) (UUID, error) {
	id, err := googleuuid.Parse(s)
	if err != nil {
		return Nil, apperror.ErrValidation.WithMessage("無効な UUID 形式です")
	}

	return id, nil
}

// Validate は文字列が有効な UUID 形式かどうかを検証する
func Validate(s string) error {
	_, err := Parse(s)
	return err
}

// MustParse は文字列を UUID に変換する（パースに失敗した場合は panic）
// テストコードでのみ使用すること
func MustParse(s string) UUID {
	return googleuuid.MustParse(s)
}
