package service

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDisplayNameToUsername(t *testing.T) {
	tests := []struct {
		name        string
		displayName string
		want        string
	}{
		{
			name:        "半角スペースをアンダースコアに変換",
			displayName: "John Doe",
			want:        "John_Doe",
		},
		{
			name:        "全角スペースをアンダースコアに変換",
			displayName: "田中　太郎",
			want:        "田中_太郎",
		},
		{
			name:        "連続する半角スペースを1つのアンダースコアに圧縮",
			displayName: "John  Doe",
			want:        "John_Doe",
		},
		{
			name:        "連続する全角スペースを1つのアンダースコアに圧縮",
			displayName: "田中　　太郎",
			want:        "田中_太郎",
		},
		{
			name:        "混合スペースを1つのアンダースコアに圧縮",
			displayName: "John 　Doe",
			want:        "John_Doe",
		},
		{
			name:        "前後の空白を削除",
			displayName: "  John Doe  ",
			want:        "John_Doe",
		},
		{
			name:        "スペースがない場合はそのまま",
			displayName: "JohnDoe",
			want:        "JohnDoe",
		},
		{
			name:        "複数の単語",
			displayName: "John Middle Doe",
			want:        "John_Middle_Doe",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := displayNameToUsername(tt.displayName)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAppendRandomSuffix(t *testing.T) {
	t.Run("ユーザー名にランダムなサフィックスを付与する", func(t *testing.T) {
		username := "testuser"
		result := appendRandomSuffix(username)

		// 形式が "testuser_数字" であることを確認
		pattern := regexp.MustCompile(`^testuser_\d+$`)
		assert.True(t, pattern.MatchString(result), "結果は 'testuser_数字' の形式であるべき: %s", result)
	})

	t.Run("サフィックスは0から9999の範囲", func(t *testing.T) {
		username := "user"
		pattern := regexp.MustCompile(`^user_(\d{1,4})$`)

		// 複数回実行して形式を確認
		for i := 0; i < 100; i++ {
			result := appendRandomSuffix(username)
			assert.True(t, pattern.MatchString(result), "サフィックスは1〜4桁の数字であるべき: %s", result)
		}
	})
}
