# Anycast Backend ドキュメント

このディレクトリには Anycast Backend の設計ドキュメントが含まれています。

## ディレクトリ構成

| ディレクトリ | 説明 |
|--------------|------|
| [specs/](./specs/) | 仕様ドキュメント（ドメインモデル、DB 設計、システム設定） |
| [api/](./api/) | API 設計ドキュメント |
| [adr/](./adr/) | Architecture Decision Records |

## 設計の流れ

本プロジェクトではドメインモデル駆動で設計を行います。

```
1. ドメインモデル設計（specs/domain-model.md）
   ↓
2. API 設計（api/）
   ↓
3. DB 設計（specs/database.md）
```

## ドキュメント更新ルール

- ドメインモデルの変更時は `domain-model.md` → `api/` → `database.md` の順で更新
- ADR を追加した際は `adr/README.md` の一覧にも追記
- DB のみの変更（インデックス追加など）は `database.md` のみ更新可
