# Anycast Backend - Claude Code 向けガイド

## ドキュメント

| ファイル | 説明 |
|----------|------|
| [docs/specs/domain-model.md](docs/specs/domain-model.md) | 仕様書（DDD ベースのドメインモデル定義） |
| [docs/specs/database.md](docs/specs/database.md) | データベース設計 |
| [docs/specs/system.md](docs/specs/system.md) | システム設定（タイムアウト、外部サービス連携など） |
| [docs/specs/script-generate-api.md](docs/specs/script-generate-api.md) | 台本生成 API 詳細設計 |
| [docs/api/README.md](docs/api/README.md) | API 設計 |
| [docs/adr/](docs/adr/) | Architecture Decision Records |

### 設計アプローチ

本プロジェクトでは **ドメインモデル駆動** で設計を行う。

```
ドメインモデル設計（specs/domain-model.md） → API 設計（api/） → DB 設計（specs/database.md）
```

- 新しい機能を追加する際は、まず domain-model.md のドメインモデルを設計する
- DB スキーマや API は、ドメインモデルを永続化・公開するための手段として設計する
- ドメインモデルの変更時は上記の順序でドキュメントを更新する

### 外部ドキュメント

| サービス | リンク |
|----------|--------|
| Gemini TTS (Vertex AI) | https://ai.google.dev/gemini-api/docs/audio |
| OpenAI API | https://platform.openai.com/docs |

## ドキュメント管理ルール

- README.md から読み取れる情報（技術スタック、ディレクトリ構成、コマンドなど）は CLAUDE.md に重複して記載しない
- ディレクトリ構成、技術スタック、バージョンなどプロジェクトの基本情報が変わった際は、README.md と CLAUDE.md の両方を更新する
- ADR を追加した際は `docs/adr/README.md` の一覧にも追記する

## 重要事項

- **Swagger を生成・更新する際は必ず `make swagger` を使用する**（`swag init` を直接実行しない）

## 開発規約

### 基本姿勢

- ユーザーの指示であっても、設計として良くないものや一般的でないものがあれば、修正を実行する前に確認を入れる
- 常にメンテナビリティやテスタビリティを意識した実装を心がける

### テスト

- 外部依存のないユニットテストは、実装時に必ず作成する

### コーディング規約

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
- バリデーションエラーは `internal/handler/validation.go` の `formatValidationError` 関数で日本語化する
  - 新しいフィールドを追加した場合は `fieldNameMap` にマッピングを追加する
  - 新しいバリデーションタグを使用する場合は `translateFieldError` 関数にケースを追加する

### ディレクトリ構成

- [golang-standards/project-layout](https://github.com/golang-standards/project-layout) を参考にする

### Git / GitHub

- ユーザーから指示があるまでコミットやプッシュを行わない（勝手にプッシュしない）
- ブランチを新規作成する際は、必ずユーザーに確認を取ってから作成する（勝手にブランチを切らない）
- コミット前に `make fmt` でフォーマットを実行する
- コミット前に `make lint` で静的解析を実行する
- PR 作成時は `.github/PULL_REQUEST_TEMPLATE.md` をテンプレートとして使用する

## 用語

| 英語 | 日本語 |
|------|--------|
| Script | 台本 |

## 学習事項

このセクションには、実装中にユーザーから指摘された内容のうち、今後の実装に役立つものを抽象化して記載する。具体的なケースではなく、同様の状況に適用できる一般的なルールとして記述すること。

### ファイル・ディレクトリ管理

- 新しいディレクトリを作成する際、空のままでも Git で管理する必要がある場合は `.gitkeep` を追加する

### 環境変数

- 環境変数を追加・変更した際は `.env.example` も更新する

### パッケージ管理

- **ライブラリ**（コードから import するもの）: `go.mod` で管理（`go get`）
- **CLI ツール**（ターミナルから実行するもの）: `.mise.toml` で管理（`mise install`）
- 依存パッケージを追加・削除した後は `go mod tidy` を実行して依存関係を整理する

### ライブラリ選定

- 標準ライブラリで対応可能でも、より便利なサードパーティライブラリがあればそちらを優先する
- Go コミュニティで広く使われているライブラリを優先的に選択する
- 新しいライブラリを導入する際は ADR を作成して決定理由を記録する

### API 実装の流れ

新しい API エンドポイントを実装する際は、以下の順序で作業する。

1. **Request DTO 追加** - `internal/dto/request/` に新しいリクエスト構造体を追加
2. **Repository** - `internal/repository/` に必要なメソッドを追加（インターフェースと実装）
3. **Service** - `internal/service/` にビジネスロジックを追加（インターフェースと実装）
4. **Handler** - `internal/handler/` にハンドラーメソッドを追加（Swagger コメント含む）
5. **Router** - `internal/router/router.go` にエンドポイントを追加
6. **DI Container** - `internal/di/container.go` の依存関係を更新（必要な場合）
7. **テスト** - モックの更新とテスト実行（`go test ./...`）
8. **ドキュメント更新**
   - `make swagger` で Swagger ドキュメントを再生成
   - `http/` ディレクトリ内の対応する `.http` ファイルを更新
   - `docs/api/README.md` の実装欄を ✅ に更新

### API ドキュメント

- ハンドラー（`internal/handler/`）を追加・変更した際は `make swagger` で Swagger ドキュメントを再生成する
- API を作成・更新した際は `http/` ディレクトリ内の対応する `.http` ファイルも作成・更新する

### DTO

- レスポンス DTO で常に値が存在するフィールドには `validate:"required"` タグを付ける
  - Swagger 生成時に `required` として出力され、フロントエンドの型がオプショナルにならない
  - `binding:"required"` はリクエストのバリデーション用なので、レスポンスには使用しない
- ポインタ型（`*string` など）や省略可能なフィールドには `validate:"required"` を付けない
- ポインタ型で `null` を返すフィールド（`omitempty` なし）には `extensions:"x-nullable"` タグを付ける
  - TypeScript の型が `string | null` のように nullable として生成される
  - `omitempty` があるフィールドは JSON から除外されるため不要
- リレーション操作（既存の紐づけ / 新規作成）を含むリクエストは `connect` / `create` パターンを使用する
  ```go
  type XxxInput struct {
      Connect []ConnectXxxInput `json:"connect"`
      Create  []CreateXxxInput  `json:"create"`
  }
  ```

### マイグレーション

- スキーマを変更した際は `docs/database.md` も更新する

### pkg ユーティリティ

- `internal/pkg/` 配下のユーティリティを積極的に使用する
- `github.com/google/uuid` の代わりに `internal/pkg/uuid` を使用する（統一されたエラーハンドリングのため）
- 汎用的な処理は `internal/pkg/` にまとめ、テストを必ず実装する

### GORM

- **プリロードされたリレーションの更新時は、リレーションフィールドを nil にクリアする**
  - `FindByID` 等でプリロードされたエンティティの外部キー（例: `ArtworkID`）を変更する際、対応するリレーションフィールド（例: `Artwork`）も `nil` に設定する
  - これをしないと、`Save` 時に古いリレーションが残り、外部キーの変更が反映されない
  ```go
  // 悪い例
  channel.ArtworkID = &newArtworkID
  repo.Update(ctx, channel)  // Artwork リレーションが古いまま → 更新されない

  // 良い例
  channel.ArtworkID = &newArtworkID
  channel.Artwork = nil  // リレーションをクリア
  repo.Update(ctx, channel)  // 正しく更新される
  ```

### ログ

| レベル | 用途 | 自動追加 |
|--------|------|:--------:|
| `Debug` | 開発時のデバッグ情報 | ✗ |
| `Info` | 運用監視用の情報（リクエストログなど） | ✗ |
| `Warn` | 注意が必要な状況（認証失敗、不正リクエストなど） | ◯ |
| `Error` | 本番環境で Slack 等に通知すべき重大なエラー | ◯ |

- ログメッセージは日本語で記載する
