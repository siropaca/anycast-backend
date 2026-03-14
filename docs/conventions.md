# 実装パターン・規約集

コード実装時に参照する規約と頻出パターンをまとめたドキュメント。

---

## 禁止事項

- `swag init` を直接実行しない（必ず `make swagger` を使う）
- `github.com/google/uuid` を直接使わない（internal/pkg/uuid を使う）
- `time.Now()` で DB タイムスタンプを設定しない（`time.Now().UTC()` を使う）
- Go 標準の `log` パッケージを使わない（internal/pkg/logger を使う）
- Go のコメントに `@param` / `@returns` などの JSDoc スタイルのタグを使わない（swag が Swagger アノテーションとして誤解釈しビルドエラーになる）

---

## コーディング規約

- Go の標準的なコーディング規約に従う
- `gofmt` でフォーマットを統一する
- エラーハンドリングは適切に行う
- ソースコードには日本語でコメントを残す（細かすぎず適切に）
- すべての関数には、その関数が何をするかを説明するコメントを追加する
- エクスポートされるシンボル（型・関数・メソッド）のコメントは、Go の標準スタイルに従いシンボル名から始める
  - 例: `// New は〜`、`// Client は〜のインターフェース`
  - 非エクスポートシンボル（小文字始まり）やコードブロック内コメントは対象外
- `interface{}` ではなく `any` を使用する
- エラーの型チェックには型アサーションではなく `errors.As` を使用する
- 標準ライブラリの新しいパッケージ（`slices`, `maps`, `cmp` など）を積極的に活用する
- エラーコードやエラーメッセージは日本語で記載する
- ログメッセージは英語で記載する（エラーメッセージは日本語のまま）
- ディレクトリ構成は [golang-standards/project-layout](https://github.com/golang-standards/project-layout) を参考にする

---

## バリデーション

- バリデーションエラーは internal/handler/validation.go の `formatValidationError` 関数で日本語化する
  - 新しいフィールドを追加した場合は `fieldNameMap` にマッピングを追加する
  - 新しいバリデーションタグを使用する場合は `translateFieldError` 関数にケースを追加する

---

## DTO

### リクエスト DTO

- `binding` タグでバリデーション
- クエリパラメータ: `form` タグを使用
- nullable な更新フィールド: `optional.Field[T]` を使用（internal/pkg/optional）
- リレーション操作（既存の紐づけ / 新規作成）を含むリクエストは `connect` / `create` パターンを使用する

```go
type XxxInput struct {
    Connect []ConnectXxxInput `json:"connect"`
    Create  []CreateXxxInput  `json:"create"`
}
```

### レスポンス DTO

- 常に値が存在するフィールドには `validate:"required"` タグを付ける
  - Swagger 生成時に `required` として出力され、フロントエンドの型がオプショナルにならない
  - `binding:"required"` はリクエストのバリデーション用なので、レスポンスには使用しない
- ポインタ型（`*string` など）や省略可能なフィールドには `validate:"required"` を付けない
- ポインタ型で `null` を返すフィールド（`omitempty` なし）には `extensions:"x-nullable"` タグを付ける
  - TypeScript の型が `string | null` のように nullable として生成される
  - `omitempty` があるフィールドは JSON から除外されるため不要

---

## GORM

### プリロードされたリレーションの更新

`FindByID` 等でプリロードされたエンティティの外部キー（例: `ArtworkID`）を変更する際、対応するリレーションフィールド（例: `Artwork`）も `nil` に設定する。これをしないと、`Save` 時に古いリレーションが残り、外部キーの変更が反映されない。

```go
// 悪い例
channel.ArtworkID = &newArtworkID
repo.Update(ctx, channel)  // Artwork リレーションが古いまま → 更新されない

// 良い例
channel.ArtworkID = &newArtworkID
channel.Artwork = nil  // リレーションをクリア
repo.Update(ctx, channel)  // 正しく更新される
```

### タイムスタンプ

DB に保存するタイムスタンプには `time.Now().UTC()` を使用する。

`time.Now()` はローカルタイムゾーン（例: JST）の時刻を返すが、PostgreSQL の `TIMESTAMP` 型（タイムゾーンなし）にそのまま保存すると、読み出し時に UTC として解釈されてタイムゾーン分ずれる。GORM の自動設定（`CreatedAt` / `UpdatedAt`）は内部的に UTC を使うため問題ないが、手動設定する `StartedAt` / `CompletedAt` 等は明示的に `.UTC()` を付ける。

```go
// 悪い例
now := time.Now()
job.StartedAt = &now  // JST がそのまま UTC として保存される

// 良い例
now := time.Now().UTC()
job.StartedAt = &now  // 正しい UTC が保存される
```

---

## ログ

| レベル | 用途 | 自動追加 |
|--------|------|:--------:|
| `Debug` | 開発時のデバッグ情報 | - |
| `Info` | 運用監視用の情報（リクエストログなど） | - |
| `Warn` | 注意が必要な状況（認証失敗、不正リクエストなど） | ◯ |
| `Error` | 本番環境で Slack 等に通知すべき重大なエラー | ◯ |

- Go 標準の `log` パッケージは使用せず、internal/pkg/logger を使用する
  - 致命的なエラーで終了する場合は `logger.Default().Error(...)` の後に `os.Exit(1)` を呼び出す

---

## pkg ユーティリティ

internal/pkg/ 配下のユーティリティを積極的に使用する。

| パッケージ | 用途 |
|------------|------|
| audio/ | 音声ファイル処理 |
| crypto/ | パスワードハッシュ |
| db/ | DB 接続 |
| jwt/ | JWT トークン管理 |
| logger/ | 構造化ログ |
| prompt/ | プロンプト圧縮 |
| script/ | 台本パーサー |
| token/ | トークン生成 |
| tracer/ | 台本生成トレーサー |
| uuid/ | UUID パース |
| cache/ | Redis キャッシュ |

- `github.com/google/uuid` の代わりに internal/pkg/uuid を使用する（統一されたエラーハンドリングのため）
- 汎用的な処理は internal/pkg/ にまとめ、テストを必ず実装する

---

## パッケージ管理

- **ライブラリ**（コードから import するもの）: `go.mod` で管理（`go get`）
- **CLI ツール**（ターミナルから実行するもの）: `.mise.toml` で管理（`mise install`）
- 依存パッケージを追加・削除した後は `go mod tidy` を実行して依存関係を整理する

## ライブラリ選定

- 標準ライブラリで対応可能でも、より便利なサードパーティライブラリがあればそちらを優先する
- Go コミュニティで広く使われているライブラリを優先的に選択する
- 新しいライブラリを導入する際は ADR を作成して決定理由を記録する

---

## マイグレーション

- スキーマを変更した際は docs/specs/database.md も更新する

## 環境変数

- 環境変数を追加・変更した際は .env.example も更新する

## ファイル・ディレクトリ管理

- 新しいディレクトリを作成する際、空のままでも Git で管理する必要がある場合は `.gitkeep` を追加する
