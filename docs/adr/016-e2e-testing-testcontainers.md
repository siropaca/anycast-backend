# ADR-016: E2E テスト: httptest + testcontainers

## ステータス

Proposed

## コンテキスト

現在、API のテストはユニットテスト（モックを使用）と手動テスト（`.http` ファイル）で行っている。
しかし、以下の課題がある：

1. **統合テストの欠如**: Repository → Service → Handler の連携が実環境で正しく動作するか検証できていない
2. **DB との結合テスト不足**: SQL クエリや GORM のプリロードが期待通り動作するか確認できていない
3. **手動テストの限界**: `.http` ファイルでのテストは再現性がなく、CI で自動化できない
4. **リグレッション検知**: 変更による既存機能への影響を自動的に検知する仕組みがない

## 決定

E2E テストのフレームワークとして **httptest + testcontainers-go** を採用する。

## 選択肢

### 選択肢 1: httptest + testcontainers-go

```go
func TestCreateScriptLine_E2E(t *testing.T) {
    ctx := context.Background()

    // PostgreSQL コンテナを起動
    postgres, _ := postgres.Run(ctx, "postgres:15")
    defer postgres.Terminate(ctx)

    // アプリケーションのセットアップ
    db := setupDB(postgres.ConnectionString())
    router := di.SetupRouter(db)

    // テスト実行
    w := httptest.NewRecorder()
    req := httptest.NewRequest("POST", "/api/v1/.../script/lines", body)
    router.ServeHTTP(w, req)

    assert.Equal(t, http.StatusCreated, w.Code)
}
```

- メリット
  - Go のエコシステムで完結（外部ツール不要）
  - 各テストが独立した DB コンテナを持てる（テスト間の干渉なし）
  - CI（GitHub Actions）で簡単に実行可能
  - Docker がインストールされていれば環境構築不要
  - テストコードと本番コードが同じ言語
- デメリット
  - コンテナ起動に時間がかかる（数秒〜十数秒）
  - Docker が必要

### 選択肢 2: テスト専用 DB（ローカル/CI）

```go
// +build integration

func TestAPI(t *testing.T) {
    db := connectTestDB(os.Getenv("TEST_DATABASE_URL"))
    // ...
}
```

- メリット
  - コンテナ起動のオーバーヘッドがない
  - シンプルなセットアップ
- デメリット
  - テスト前に DB を準備する必要がある
  - テスト間でデータが干渉する可能性
  - ローカル環境と CI で設定を揃える必要がある

### 選択肢 3: Newman（Postman CLI）

```bash
newman run collection.json --environment env.json
```

- メリット
  - Postman を使っている場合は資産を再利用できる
  - GUI でテストケースを作成可能
- デメリット
  - Go のエコシステム外
  - テストと実装が別管理になる
  - 複雑なセットアップ/検証が書きにくい

### 選択肢 4: hurl

```hurl
POST http://localhost:8081/api/v1/.../script/lines
Authorization: Bearer {{token}}
{
  "speakerId": "...",
  "text": "テスト"
}
HTTP 201
```

- メリット
  - シンプルな構文
  - CI で実行可能
- デメリット
  - Go 以外のツールが必要
  - 複雑な検証やセットアップが難しい
  - DB 状態の検証ができない

## 理由

1. **一貫性**: Go + testify で統一され、既存のユニットテストと同じ書き方ができる
2. **独立性**: testcontainers によりテストごとに独立した DB 環境を持てる
3. **CI 親和性**: GitHub Actions で追加設定なしに動作する
4. **保守性**: テストコードがアプリケーションコードと同じリポジトリ・言語で管理できる
5. **信頼性**: 実際の PostgreSQL を使うため、本番に近い環境でテストできる

## 結果

- `github.com/testcontainers/testcontainers-go` を依存関係に追加
- `e2e/` ディレクトリを作成し、E2E テストを配置
- Makefile に `test-e2e` ターゲットを追加
- CI（GitHub Actions）に E2E テストのジョブを追加

### ディレクトリ構成

```
project/
├── internal/
│   ├── handler/
│   │   ├── script_line.go
│   │   └── script_line_test.go      # ユニットテスト（モック）
│   └── ...
├── e2e/                              # E2E テスト
│   ├── setup_test.go                 # 共通セットアップ
│   ├── script_line_test.go
│   └── auth_test.go
└── Makefile
```

### Makefile

```makefile
# ユニットテストのみ
test:
	go test ./internal/...

# E2E テストのみ
test-e2e:
	go test ./e2e/...

# 全テスト
test-all:
	go test ./...
```
