# ドキュメント

## ディレクトリ構成

| パス | 内容 | 概要 |
|------|------|------|
| [domain-model/INDEX.md](domain-model/INDEX.md) | ドメインモデル | 集約・エンティティ・値オブジェクトの定義 |
| [specs/INDEX.md](specs/INDEX.md) | 仕様書 | DB 設計、システム設定、非同期 API 詳細設計 |
| [api/INDEX.md](api/INDEX.md) | API 設計 | REST API 一覧、各エンドポイントの詳細仕様、共通仕様、エラーコード |
| [adr/INDEX.md](adr/INDEX.md) | ADR | 技術選定の意思決定記録（20件） |
| [conventions.md](conventions.md) | 実装パターン | コーディング規約、DTO、GORM、ログ等の実装パターン集 |
| [testing.md](testing.md) | テストガイド | ユニットテスト、API 手動検証の手順 |
| [definition-of-done.md](definition-of-done.md) | DoD | タスク完了の判断基準チェックリスト |
| [ubiquitous-language.md](ubiquitous-language.md) | ユビキタス言語集 | プロジェクト全体で統一する用語の定義 |

## ドキュメント間の依存関係

ドメインモデルを起点に、各ドキュメントは以下の依存関係を持つ。変更時は上流から順に更新する。

```
domain-model/  ──→  api/         ──→  specs/database.md
 (ドメイン定義)      (API 仕様)        (DB スキーマ)
      │                                      ↑
      └──────────────────────────────────────┘
                   直接参照（属性 ↔ カラム対応）

conventions.md ← 実装時に参照（全レイヤー共通）
testing.md     ← テスト実装時に参照
adr/           ← 新規ライブラリ導入時に作成、技術選定の理由確認時に参照
```

## ドキュメント更新ルール

- ドメインモデルの変更時は `domain-model/` → `api/` → `specs/database.md` の順で更新
- ADR を追加した際は `adr/INDEX.md` の一覧にも追記
- DB のみの変更（インデックス追加など）は `specs/database.md` のみ更新可
- 新しいサブディレクトリを作成した場合は INDEX.md を設置する
- ドキュメント内のリンクは相対パスで記述する
