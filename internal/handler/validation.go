package handler

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// フィールド名の日本語マッピング
var fieldNameMap = map[string]string{
	// 共通
	"ID":   "ID",
	"Name": "名前",

	// 認証
	"Email":           "メールアドレス",
	"Password":        "パスワード",
	"DisplayName":     "表示名",
	"ProviderUserID":  "プロバイダーユーザー ID",
	"AccessToken":     "アクセストークン",
	"RefreshToken":    "リフレッシュトークン",
	"ExpiresAt":       "有効期限",
	"CurrentPassword": "現在のパスワード",
	"NewPassword":     "新しいパスワード",
	"Username":        "ユーザー名",

	// ユーザープロフィール
	"Bio":           "自己紹介",
	"AvatarImageID": "アバター画像 ID",
	"HeaderImageID": "ヘッダー画像 ID",

	// チャンネル・エピソード
	"Title":              "タイトル",
	"Description":        "説明",
	"UserPrompt":         "ユーザープロンプト",
	"CategoryID":         "カテゴリ ID",
	"CategorySlug":       "カテゴリスラッグ",
	"ArtworkImageID":     "アートワーク画像 ID",
	"DefaultBgmID":       "デフォルト BGM ID",
	"DefaultSystemBgmID": "デフォルトシステム BGM ID",
	"Characters":         "キャラクター",
	"PublishedAt":        "公開日時",

	// キャラクター
	"Persona":  "ペルソナ",
	"AvatarID": "アバター ID",
	"VoiceID":  "ボイス ID",

	// BGM・音声
	"AudioID":     "音声 ID",
	"BgmID":       "BGM ID",
	"SystemBgmID": "システム BGM ID",
	// 台本
	"Prompt":          "プロンプト",
	"DurationMinutes": "再生時間（分）",
	"Text":            "テキスト",
	"Emotion":         "感情",
	"SpeakerID":       "スピーカー ID",
	"AfterLineID":     "挿入位置の行 ID",
	"LineIDs":         "行 ID リスト",
	"WithEmotion":     "感情付与",

	// リアクション
	"ReactionType": "リアクションタイプ",

	// 再生リスト
	"PlaylistIDs": "再生リスト ID リスト",

	// 検索
	"Q": "検索キーワード",

	// その他
	"Status": "ステータス",
}

// formatValidationError はバリデーションエラーを日本語メッセージに変換する
func formatValidationError(err error) string {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		var messages []string
		for _, fe := range ve {
			messages = append(messages, translateFieldError(fe))
		}
		return strings.Join(messages, ", ")
	}
	return err.Error()
}

// translateFieldError は個別のフィールドエラーを日本語に変換する
func translateFieldError(fe validator.FieldError) string {
	fieldName := getFieldName(fe.Field())

	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s は必須です", fieldName)
	case "max":
		return fmt.Sprintf("%s は %s 文字以内で入力してください", fieldName, fe.Param())
	case "min":
		return fmt.Sprintf("%s は %s 文字以上で入力してください", fieldName, fe.Param())
	case "uuid":
		return fmt.Sprintf("%s は有効な UUID 形式で入力してください", fieldName)
	case "email":
		return fmt.Sprintf("%s は有効なメールアドレス形式で入力してください", fieldName)
	case "oneof":
		return fmt.Sprintf("%s は許可された値（%s）のいずれかを指定してください", fieldName, fe.Param())
	default:
		return fmt.Sprintf("%s の形式が正しくありません", fieldName)
	}
}

// getFieldName はフィールド名を日本語に変換する
func getFieldName(field string) string {
	if name, ok := fieldNameMap[field]; ok {
		return name
	}
	return field
}
