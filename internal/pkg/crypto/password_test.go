package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPasswordHasher_Hash(t *testing.T) {
	hasher := NewPasswordHasher()

	t.Run("パスワードをハッシュ化できる", func(t *testing.T) {
		password := "testpassword123"

		hashed, err := hasher.Hash(password)

		require.NoError(t, err)
		assert.NotEmpty(t, hashed)
		assert.NotEqual(t, password, hashed)
	})

	t.Run("同じパスワードでも異なるハッシュ値が生成される", func(t *testing.T) {
		password := "testpassword123"

		hashed1, err := hasher.Hash(password)
		require.NoError(t, err)

		hashed2, err := hasher.Hash(password)
		require.NoError(t, err)

		assert.NotEqual(t, hashed1, hashed2)
	})
}

func TestPasswordHasher_Compare(t *testing.T) {
	hasher := NewPasswordHasher()

	t.Run("正しいパスワードで検証が成功する", func(t *testing.T) {
		password := "testpassword123"
		hashed, err := hasher.Hash(password)
		require.NoError(t, err)

		err = hasher.Compare(hashed, password)

		assert.NoError(t, err)
	})

	t.Run("誤ったパスワードで検証が失敗する", func(t *testing.T) {
		password := "testpassword123"
		wrongPassword := "wrongpassword"
		hashed, err := hasher.Hash(password)
		require.NoError(t, err)

		err = hasher.Compare(hashed, wrongPassword)

		assert.Error(t, err)
	})

	t.Run("空のパスワードで検証が失敗する", func(t *testing.T) {
		password := "testpassword123"
		hashed, err := hasher.Hash(password)
		require.NoError(t, err)

		err = hasher.Compare(hashed, "")

		assert.Error(t, err)
	})
}
