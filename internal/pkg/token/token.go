package token

import (
	"crypto/rand"
	"encoding/base64"
)

const tokenBytes = 32

// Generate はランダムなリフレッシュトークンを生成する
func Generate() (string, error) {
	b := make([]byte, tokenBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(b), nil
}
