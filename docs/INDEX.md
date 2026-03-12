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

## 設計の流れ

本プロジェクトではドメインモデル駆動で設計を行う。

```
1. ドメインモデル設計（docs/domain-model/）
   ↓
2. API 設計（docs/api/）
   ↓
3. DB 設計（docs/specs/database.md）
```

## ドキュメント更新ルール

- ドメインモデルの変更時は `docs/domain-model/` → `docs/api/` → `docs/specs/database.md` の順で更新
- ADR を追加した際は `docs/adr/INDEX.md` の一覧にも追記
- DB のみの変更（インデックス追加など）は `docs/specs/database.md` のみ更新可
