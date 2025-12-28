package service

import (
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
