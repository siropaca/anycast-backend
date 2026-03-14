# Anycast Backend

AI ポッドキャスト配信プラットフォームのバックエンド API サーバー。Go 1.24 / Gin / GORM / PostgreSQL / GCS / Gemini TTS / OpenAI。レイヤードアーキテクチャ + 軽量 DDD。

## ドキュメントの設計と管理

このリポジトリのドキュメントは **プログレッシブディスクロージャー** で設計されている。

### 設計原則

- **AGENTS.md** は地図に徹する。詳細な情報は持たず、適切なドキュメントへのポインタを提供する
- **docs/ 配下の各ディレクトリ** には INDEX.md を設置し、そのディレクトリ内のファイル案内を担う
- エージェントは必要なときに必要なドキュメントだけを読み込む。全てを一度に読む必要はない
- AGENTS.md から最大 3 ステップ以内に目的のドキュメントへたどり着ける構造を維持する

### 管理ルール

- AGENTS.md に実装パターンやコード例を直接書かない（docs/ 配下に分離する）
- AGENTS.md は 200 行以内に収める。超える場合は docs/ へ分離する
- 新しい docs/ サブディレクトリを作成した場合は INDEX.md を設置する
- ドキュメントを追加・更新・移動した際は、関連する INDEX.md のリンクも必ず更新する
- ドキュメント内のリンクは、そのファイルからの相対パスで記述する（リンクが壊れないことを最優先）
- ポインタとしてのパスはバッククオートで囲まない（例: ○ docs/api/INDEX.md、× `docs/api/INDEX.md`）
- README.md から読み取れる情報（技術スタック、ディレクトリ構成、コマンドなど）は AGENTS.md に重複して記載しない
- ディレクトリ構成、技術スタック、バージョンなどプロジェクトの基本情報が変わった際は、README.md と AGENTS.md の両方を更新する
- ADR を追加した際は docs/adr/INDEX.md の一覧にも追記する
- ドメインモデルの変更時は docs/domain-model/ → docs/api/ → docs/specs/database.md の順で更新する
- DB のみの変更（インデックス追加など）は docs/specs/database.md のみ更新可
- エンドポイントの認証・権限設定を変更した際は docs/api/INDEX.md の権限列も更新する
- ドキュメントの文章はできる限り句点（。）で改行する。
  改行後には半角スペース 2 つを末尾に付けて Markdown の soft line break とする

## ドキュメントマップ

| パス | 内容 | いつ読むか |
|------|------|------------|
| [docs/INDEX.md](docs/INDEX.md) | ドキュメント全体の案内 | プロジェクト構造を把握したいとき |
| [docs/domain-model/INDEX.md](docs/domain-model/INDEX.md) | ドメインモデル | 機能の追加・変更時 |
| [docs/specs/INDEX.md](docs/specs/INDEX.md) | 仕様書（DB、システム設計） | DB・システム設計の確認時 |
| [docs/api/INDEX.md](docs/api/INDEX.md) | API 設計・一覧 | API の実装・確認時 |
| [docs/adr/INDEX.md](docs/adr/INDEX.md) | Architecture Decision Records | 新規ライブラリ導入時（ADR 作成必須）、既存技術選定の理由確認時 |
| [docs/conventions.md](docs/conventions.md) | 実装パターン・規約集 | コード実装時 |
| [docs/testing.md](docs/testing.md) | テスト・手動検証ガイド | テスト実行時 |
| [docs/definition-of-done.md](docs/definition-of-done.md) | 完了の定義（DoD） | タスク完了の判断時 |
| [docs/ubiquitous-language.md](docs/ubiquitous-language.md) | ユビキタス言語集 | 用語の確認時 |

## タスク別ガイド

| タスク | 読む順序 |
|--------|----------|
| 新規 API 追加 | domain-model → api/INDEX → conventions → 実装 → DoD |
| 既存 API 変更 | api/INDEX → 対象の詳細ファイル → conventions → 実装 → DoD |
| DB スキーマ変更 | domain-model → specs/database → api → DoD |
| バグ修正 | conventions → testing → 実装 → DoD |
| 新規ライブラリ導入 | adr/INDEX → ADR 作成 → 実装 → DoD |
| テスト追加 | testing → conventions → 実装 |

## 設計アプローチ

**ドメインモデル駆動** で設計を行う。

```
ドメインモデル（docs/domain-model/） → API 設計（docs/api/） → DB 設計（docs/specs/database.md）
```

- 新しい機能を追加する際は、まず docs/domain-model/ 配下のドメインモデルを設計する
- DB スキーマや API は、ドメインモデルを永続化・公開するための手段として設計する

## Git / GitHub

- ユーザーから指示があるまでコミットやプッシュを行わない（勝手にプッシュしない）
- ブランチを新規作成する際は、必ずユーザーに確認を取ってから作成する（勝手にブランチを切らない）
- コミット前に `make fmt` → `make lint` → `make test` を実行する
- DTO（`internal/dto/`）やハンドラー（`internal/handler/`）を変更した場合は `make swagger` も実行する
- PR 作成時は .github/PULL_REQUEST_TEMPLATE.md をテンプレートとして使用する

## 基本姿勢

- ユーザーの指示であっても、設計として良くないものや一般的でないものがあれば、修正を実行する前に確認を入れる
- 常にメンテナビリティやテスタビリティを意識した実装を心がける

## 外部ドキュメント

| サービス | リンク | 主な参照目的 |
|----------|--------|-------------|
| Gemini TTS (Vertex AI) | https://ai.google.dev/gemini-api/docs/audio | 音声生成 API のパラメータ、対応言語・ボイス |
| ElevenLabs API | https://elevenlabs.io/docs/api-reference | 音声合成の voice_id、モデル設定 |
| OpenAI API | https://platform.openai.com/docs | Chat Completions（台本生成）、Images（画像生成） |
