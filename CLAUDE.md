# Anycast Backend - Claude Code 向けガイド

## ドキュメント

| ファイル | 説明 |
|----------|------|
| [doc/specification.md](doc/specification.md) | サービス仕様書 |
| [doc/database.md](doc/database.md) | データベース設計 |
| [doc/api.md](doc/api.md) | API 設計 |

## ドキュメント管理ルール

- README.md から読み取れる情報（技術スタック、ディレクトリ構成、コマンドなど）は CLAUDE.md に重複して記載しない
- ディレクトリ構成、技術スタック、バージョンなどプロジェクトの基本情報が変わった際は、README.md と CLAUDE.md の両方を更新する

## 開発規約

### コーディング規約

- Go の標準的なコーディング規約に従う
- `gofmt` でフォーマットを統一する
- エラーハンドリングは適切に行う

### ディレクトリ構成

- [golang-standards/project-layout](https://github.com/golang-standards/project-layout) を参考にする

### コミット・プッシュ

- ユーザーから指示があるまでコミットやプッシュを行わない
- コミット前に `make fmt` でフォーマットを実行する
- コミット前に `make lint` で静的解析を実行する

## 学習事項

このセクションには、実装中にユーザーから指摘された内容のうち、今後の実装に役立つものを抽象化して記載する。具体的なケースではなく、同様の状況に適用できる一般的なルールとして記述すること。

### ファイル・ディレクトリ管理

- 新しいディレクトリを作成する際、空のままでも Git で管理する必要がある場合は `.gitkeep` を追加する

### パッケージ管理

- **ライブラリ**（コードから import するもの）: `go.mod` で管理（`go get`）
- **CLI ツール**（ターミナルから実行するもの）: `.mise.toml` で管理（`mise install`）
