// 開発環境用の JWT トークンを生成するスクリプト
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
)

// テストユーザーの ID（seeds/001_users.sql と一致）
const defaultUserID = "8def69af-dae9-4641-a0e5-100107626933"

func main() {
	_ = godotenv.Load() //nolint:errcheck // .env ファイルがなくてもエラーにしない

	jwtSecret := os.Getenv("AUTH_SECRET")
	if jwtSecret == "" {
		fmt.Fprintln(os.Stderr, "Error: AUTH_SECRET is not set")
		os.Exit(1)
	}

	// ユーザー ID（引数で指定可能）
	userID := defaultUserID
	if len(os.Args) > 1 {
		userID = os.Args[1]
	}

	// トークンを生成（有効期限: 1時間）
	claims := jwt.MapClaims{
		"sub": userID,
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(1 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to sign token: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(tokenString)
}
