package uuid

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/siropaca/anycast-backend/internal/apperror"
)

func TestParse(t *testing.T) {
	t.Run("有効な UUID をパースできる", func(t *testing.T) {
		validUUID := "550e8400-e29b-41d4-a716-446655440000"

		id, err := Parse(validUUID)

		require.NoError(t, err)
		assert.Equal(t, validUUID, id.String())
	})

	t.Run("無効な UUID はエラーになる", func(t *testing.T) {
		invalidUUID := "invalid-uuid"

		id, err := Parse(invalidUUID)

		assert.Error(t, err)
		assert.Equal(t, Nil, id)
		assert.True(t, apperror.IsCode(err, apperror.CodeValidation))
	})

	t.Run("空文字列はエラーになる", func(t *testing.T) {
		id, err := Parse("")

		assert.Error(t, err)
		assert.Equal(t, Nil, id)
		assert.True(t, apperror.IsCode(err, apperror.CodeValidation))
	})

	t.Run("ハイフンなしの UUID もパースできる", func(t *testing.T) {
		uuidWithoutHyphens := "550e8400e29b41d4a716446655440000"

		id, err := Parse(uuidWithoutHyphens)

		require.NoError(t, err)
		assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", id.String())
	})
}

func TestValidate(t *testing.T) {
	t.Run("有効な UUID は nil を返す", func(t *testing.T) {
		err := Validate("550e8400-e29b-41d4-a716-446655440000")

		assert.NoError(t, err)
	})

	t.Run("無効な UUID はエラーを返す", func(t *testing.T) {
		err := Validate("invalid-uuid")

		assert.Error(t, err)
		assert.True(t, apperror.IsCode(err, apperror.CodeValidation))
	})
}
