# Anycast Backend - Claude Code 向けガイド

## ドキュメント

| ファイル | 説明 |
|----------|------|
| [docs/specification.md](docs/specification.md) | サービス仕様書 |
| [docs/database.md](docs/database.md) | データベース設計 |
| [docs/api.md](docs/api.md) | API 設計 |
| [docs/adr/](docs/adr/) | Architecture Decision Records |

### 外部ドキュメント

| サービス | リンク |
|----------|--------|
| Google Cloud Text-to-Speech | https://cloud.google.com/text-to-speech/docs |

## ドキュメント管理ルール

- README.md から読み取れる情報（技術スタック、ディレクトリ構成、コマンドなど）は CLAUDE.md に重複して記載しない
- ディレクトリ構成、技術スタック、バージョンなどプロジェクトの基本情報が変わった際は、README.md と CLAUDE.md の両方を更新する
- ADR を追加した際は `docs/adr/README.md` の一覧にも追記する

## 開発規約

### コーディング規約

- Go の標準的なコーディング規約に従う
- `gofmt` でフォーマットを統一する
- エラーハンドリングは適切に行う
- ソースコードには日本語でコメントを残す（細かすぎず適切に）
- すべての関数には、その関数が何をするかを説明するコメントを追加する
- 関数コメントの先頭に関数名を書かない（例: `// New は〜` ではなく `// 〜` と書く）
- `interface{}` ではなく `any` を使用する
- エラーの型チェックには型アサーションではなく `errors.As` を使用する
- 標準ライブラリの新しいパッケージ（`slices`, `maps`, `cmp` など）を積極的に活用する

### ディレクトリ構成

- [golang-standards/project-layout](https://github.com/golang-standards/project-layout) を参考にする

### Git / GitHub

- ユーザーから指示があるまでコミットやプッシュを行わない（勝手にプッシュしない）
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

### パッケージ管理

- **ライブラリ**（コードから import するもの）: `go.mod` で管理（`go get`）
- **CLI ツール**（ターミナルから実行するもの）: `.mise.toml` で管理（`mise install`）
- 依存パッケージを追加・削除した後は `go mod tidy` を実行して依存関係を整理する
