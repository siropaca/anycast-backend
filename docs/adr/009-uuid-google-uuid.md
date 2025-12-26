# ADR-009: UUID ライブラリとして google/uuid を採用

## ステータス

Accepted

## コンテキスト

全テーブルで UUID を主キーとして使用する設計のため、UUID を扱うライブラリを選定する必要があった。

## 決定

**github.com/google/uuid** を採用する。

```go
import "github.com/google/uuid"

type Voice struct {
    ID uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
}

// 新規生成
id := uuid.New()

// パース
id, err := uuid.Parse("550e8400-e29b-41d4-a716-446655440000")
```

## 選択肢

### 選択肢 1: google/uuid

- Google がメンテナンス
- シンプルな API
- UUID v1, v4, v6, v7 をサポート
- 広く使われている

### 選択肢 2: gofrs/uuid

- google/uuid のフォーク
- 追加機能あり（UUID v2 など）
- やや複雑な API

### 選択肢 3: satori/go.uuid

- 古くから使われている
- メンテナンスが停滞
- google/uuid への移行が推奨

### 選択肢 4: 文字列で管理

- ライブラリ不要
- バリデーションが手動
- 型安全性がない

## 理由

1. **信頼性**: Google がメンテナンスしており長期的に安定
2. **シンプルさ**: 必要十分な機能
3. **互換性**: GORM との統合が容易
4. **標準的**: Go プロジェクトで最も使われている UUID ライブラリ

## 結果

- `github.com/google/uuid` を使用
- モデルの ID フィールドは `uuid.UUID` 型
- PostgreSQL では `gen_random_uuid()` でデフォルト値を生成
- JSON シリアライズは自動的に文字列形式
