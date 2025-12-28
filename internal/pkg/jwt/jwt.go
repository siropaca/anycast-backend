package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken     = errors.New("invalid token")
	ErrInvalidAlgorithm = errors.New("invalid signing algorithm")
)

// JWT クレーム
type Claims struct {
	jwt.RegisteredClaims
	UserID string `json:"-"` // Subject から設定される
}

// JWT の生成と検証を行うインターフェース
type TokenManager interface {
	// トークンを生成する
	Generate(userID string, expiration time.Duration) (string, error)
	// トークンを検証してクレームを返す
	Validate(tokenString string) (*Claims, error)
}

type tokenManager struct {
	secret []byte
}

// TokenManager の実装を返す
func NewTokenManager(secret string) TokenManager {
	return &tokenManager{
		secret: []byte(secret),
	}
}

// トークンを生成する
func (m *tokenManager) Generate(userID string, expiration time.Duration) (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   userID,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(expiration)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

// トークンを検証してクレームを返す
func (m *tokenManager) Validate(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		// 署名アルゴリズムを検証（HMAC のみ許可）
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidAlgorithm
		}
		return m.secret, nil
	})

	if err != nil {
		return nil, errors.Join(ErrInvalidToken, err)
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	// Subject から UserID を設定
	claims.UserID = claims.Subject

	return claims, nil
}
