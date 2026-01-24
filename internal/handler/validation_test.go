package handler

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

func TestFormatValidationError(t *testing.T) {
	validate := validator.New()

	tests := []struct {
		name     string
		input    any
		expected string
	}{
		{
			name: "required タグのエラー",
			input: struct {
				Name string `validate:"required"`
			}{Name: ""},
			expected: "名前 は必須です",
		},
		{
			name: "max タグのエラー",
			input: struct {
				Name string `validate:"max=10"`
			}{Name: "12345678901"},
			expected: "名前 は 10 文字以内で入力してください",
		},
		{
			name: "min タグのエラー",
			input: struct {
				Password string `validate:"min=8"`
			}{Password: "short"},
			expected: "パスワード は 8 文字以上で入力してください",
		},
		{
			name: "uuid タグのエラー",
			input: struct {
				ID string `validate:"uuid"`
			}{ID: "invalid-uuid"},
			expected: "ID は有効な UUID 形式で入力してください",
		},
		{
			name: "email タグのエラー",
			input: struct {
				Email string `validate:"email"`
			}{Email: "invalid-email"},
			expected: "メールアドレス は有効なメールアドレス形式で入力してください",
		},
		{
			name: "oneof タグのエラー",
			input: struct {
				Status string `validate:"oneof=active inactive"`
			}{Status: "unknown"},
			expected: "ステータス は許可された値（active inactive）のいずれかを指定してください",
		},
		{
			name: "複数のエラー",
			input: struct {
				Name  string `validate:"required"`
				Email string `validate:"required,email"`
			}{Name: "", Email: ""},
			expected: "名前 は必須です, メールアドレス は必須です",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate.Struct(tt.input)
			assert.Error(t, err)

			result := formatValidationError(err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatValidationError_NonValidationError(t *testing.T) {
	err := assert.AnError
	result := formatValidationError(err)
	assert.Equal(t, err.Error(), result)
}

func TestGetFieldName(t *testing.T) {
	tests := []struct {
		field    string
		expected string
	}{
		{"Name", "名前"},
		{"Email", "メールアドレス"},
		{"Password", "パスワード"},
		{"DisplayName", "表示名"},
		{"Title", "タイトル"},
		{"Description", "説明"},
		{"UserPrompt", "ユーザープロンプト"},
		{"CategoryID", "カテゴリ ID"},
		{"Characters", "キャラクター"},
		{"VoiceID", "ボイス ID"},
		{"UnknownField", "UnknownField"}, // マッピングがない場合は元の名前
	}

	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			result := getFieldName(tt.field)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTranslateFieldError_WithJapaneseFieldName(t *testing.T) {
	validate := validator.New()

	// 日本語マッピングがあるフィールドのテスト
	input := struct {
		Email string `validate:"required"`
	}{Email: ""}

	err := validate.Struct(input)
	assert.Error(t, err)

	result := formatValidationError(err)
	assert.Equal(t, "メールアドレス は必須です", result)
}

func TestTranslateFieldError_UnknownTag(t *testing.T) {
	validate := validator.New()

	// alphanum タグはマッピングにないのでデフォルトメッセージになる
	input := struct {
		Code string `validate:"alphanum"`
	}{Code: "abc-123"}

	err := validate.Struct(input)
	assert.Error(t, err)

	result := formatValidationError(err)
	assert.Equal(t, "Code の形式が正しくありません", result)
}
