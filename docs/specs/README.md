# Anycast 仕様ドキュメント

このディレクトリには Anycast Backend の仕様ドキュメントが含まれています。

## ドキュメント一覧

| ファイル | 説明 |
|----------|------|
| [domain-model.md](./domain-model.md) | ドメインモデル設計書。エンティティ、集約、値オブジェクトの定義 |
| [database.md](./database.md) | データベース設計書。ER 図、テーブル定義、制約 |
| [script-generate-async-api.md](./script-generate-async-api.md) | 台本生成 API（非同期）の詳細設計。LLM プロンプト設計、Cloud Tasks、WebSocket |
| [script-prompt-workflow.md](./script-prompt-workflow.md) | 台本生成プロンプトワークフロー仕様。多段階生成・品質検証の設計 |
| [audio-generate-async-api.md](./audio-generate-async-api.md) | 音声生成 API（非同期）の詳細設計。Cloud Tasks、TTS、WebSocket |
| [system.md](./system.md) | システム設定。タイムアウト、外部サービス設定 |

## 設計の流れ

本プロジェクトではドメインモデル駆動で設計を行います。

```
1. ドメインモデル設計（domain-model.md）
   ↓
2. API 設計（../api/）
   ↓
3. DB 設計（database.md）
```

詳細は [domain-model.md](./domain-model.md) を参照してください。
