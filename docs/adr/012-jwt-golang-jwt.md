# ADR-012: JWT ライブラリ: golang-jwt/jwt

## ステータス

Accepted

## コンテキスト

バックエンド API の認証に Bearer Token（JWT）を使用する必要がある。フロントエンドの Next Auth から発行された JWT を検証し、ユーザー認証を行うための JWT ライブラリを選定する。

## 決定

JWT ライブラリとして `github.com/golang-jwt/jwt/v5` を採用する。

## 選択肢

### 選択肢 1: golang-jwt/jwt/v5

- **メリット**
  - Go コミュニティで最も広く使用されている JWT ライブラリ
  - 元の `dgrijalva/jwt-go` の公式後継プロジェクト
  - 活発にメンテナンスされている
  - シンプルな API
  - 複数の署名アルゴリズム（HS256, RS256 など）をサポート
- **デメリット**
  - 特になし

### 選択肢 2: go-jose/go-jose

- **メリット**
  - JOSE（JSON Object Signing and Encryption）の完全な実装
  - JWE（暗号化）もサポート
- **デメリット**
  - JWT のみの用途には過剰
  - API が複雑

### 選択肢 3: lestrrat-go/jwx

- **メリット**
  - JOSE 標準の完全な実装
  - JWK, JWS, JWE, JWT すべてをサポート
- **デメリット**
  - シンプルな JWT 検証には過剰
  - 学習コストが高い

## 理由

1. **業界標準**: `golang-jwt/jwt` は Go における JWT 処理のデファクトスタンダードであり、多くのドキュメントやサンプルコードが存在する
2. **シンプルさ**: 今回の用途（Next Auth からの JWT 検証）にはシンプルな API で十分
3. **メンテナンス**: 活発に開発が続いており、セキュリティアップデートも迅速に提供される

## 結果

- `go get github.com/golang-jwt/jwt/v5` で依存関係を追加
- 認証ミドルウェアで JWT の検証ロジックを実装
- Next Auth と同じ署名アルゴリズム・シークレットを使用して検証
