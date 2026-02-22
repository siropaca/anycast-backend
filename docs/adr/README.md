# Architecture Decision Records (ADR)

このディレクトリには、技術的意思決定を記録した ADR を格納しています。

## ADR とは

ADR（Architecture Decision Record）は、アーキテクチャ上の重要な決定を記録するドキュメントです。
なぜその決定を行ったのか、どのような選択肢を検討したのかを残すことで、将来の開発者が決定の背景を理解できるようにします。

## ADR 一覧

| 番号 | タイトル | ステータス |
|------|----------|------------|
| [001](001-layered-architecture.md) | レイヤードアーキテクチャの採用 | Accepted |
| [002](002-manual-dependency-injection.md) | 手動による依存性注入 | Accepted |
| [003](003-structured-logging-with-slog.md) | slog による構造化ログ | Accepted |
| [004](004-openapi-swagger.md) | OpenAPI (Swagger) による API ドキュメント | Accepted |
| [005](005-custom-error-type.md) | カスタムエラー型 | Accepted |
| [006](006-web-framework-gin.md) | Web フレームワーク: Gin | Accepted |
| [007](007-orm-gorm.md) | ORM: GORM | Accepted |
| [008](008-migration-golang-migrate.md) | マイグレーション: golang-migrate | Accepted |
| [009](009-uuid-google-uuid.md) | UUID: google/uuid | Accepted |
| [010](010-linter-golangci-lint.md) | 静的解析: golangci-lint | Accepted |
| [011](011-testing-testify.md) | テストライブラリ: testify | Accepted |
| [012](012-jwt-golang-jwt.md) | JWT ライブラリ: golang-jwt/jwt | Accepted |
| [013](013-transaction-in-service-layer.md) | トランザクション管理は Service 層で行う | Accepted |
| [014](014-orphaned-media-cleanup-api.md) | 孤児メディアファイル削除 API | Accepted |
| [015](015-user-role-enum.md) | ユーザー権限管理に Enum 型を使用 | Accepted |
| [016](016-e2e-testing-testcontainers.md) | E2E テスト: httptest + testcontainers | Proposed |
| [017](017-audio-mixing-ffmpeg.md) | 音声ミキシングに FFmpeg を使用 | Accepted |
| [018](018-monolith-with-cloud-tasks.md) | ワーカー処理を分離せずモノリス + Cloud Tasks 構成を維持 | Proposed |
| [019](019-stt-timestamp-audio-segmentation.md) | STT タイムスタンプによる音声セグメント分割 | Accepted |

## ステータス

- **Proposed**: 提案中
- **Accepted**: 採用
- **Deprecated**: 非推奨（別の ADR で置き換え）
- **Superseded**: 別の ADR で上書き

## 新しい ADR の作成

`template.md` をコピーして、連番でファイルを作成してください。

```bash
cp docs/adr/template.md docs/adr/00X-title.md
```
