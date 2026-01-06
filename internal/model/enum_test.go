package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRole_IsAdmin(t *testing.T) {
	t.Run("admin ロールの場合は true を返す", func(t *testing.T) {
		role := RoleAdmin
		assert.True(t, role.IsAdmin())
	})

	t.Run("user ロールの場合は false を返す", func(t *testing.T) {
		role := RoleUser
		assert.False(t, role.IsAdmin())
	})

	t.Run("空の Role の場合は false を返す", func(t *testing.T) {
		role := Role("")
		assert.False(t, role.IsAdmin())
	})
}
