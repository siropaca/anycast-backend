package crypto

import (
	"golang.org/x/crypto/bcrypt"
)

// PasswordHasher はパスワードのハッシュ化と検証を行うインターフェース
type PasswordHasher interface {
	// Hash はパスワードをハッシュ化する
	Hash(password string) (string, error)
	// Compare はパスワードとハッシュ値を比較する
	Compare(hashedPassword, password string) error
}

type bcryptHasher struct {
	cost int
}

// NewPasswordHasher は bcrypt を使用した PasswordHasher を返す
func NewPasswordHasher() PasswordHasher {
	return &bcryptHasher{
		cost: bcrypt.DefaultCost,
	}
}

// Hash はパスワードをハッシュ化する
func (h *bcryptHasher) Hash(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

// Compare はパスワードとハッシュ値を比較する
func (h *bcryptHasher) Compare(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
