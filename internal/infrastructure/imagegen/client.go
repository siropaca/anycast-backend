package imagegen

import "context"

// GenerateResult は画像生成の結果を表す
type GenerateResult struct {
	Data     []byte // 画像バイナリデータ
	MimeType string // MIME タイプ（例: "image/png"）
}

// Client は画像生成クライアントのインターフェース
type Client interface {
	// Generate はテキストプロンプトから画像を生成する
	Generate(ctx context.Context, prompt string) (*GenerateResult, error)
}
