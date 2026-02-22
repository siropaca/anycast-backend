# API Keys（API キー）

JWT Bearer トークンの代替認証手段として使用できる API キーの管理。

- 平文キーは作成時に 1 度だけ返却される（再取得不可）
- DB には SHA-256 ハッシュのみ保存
- 認証方式: `X-API-Key` ヘッダーまたは `Authorization: Bearer ak_...`

---

## API キー作成

新しい API キーを発行する。

### リクエスト

```
POST /api/v1/me/api-keys
Authorization: Bearer {token}
```

```json
{
  "name": "My API Key"
}
```

| フィールド | 型 | 必須 | 説明 |
|-----------|-----|:----:|------|
| name | String | ◯ | 管理名（1〜100文字、同一ユーザー内で一意） |

### レスポンス

**201 Created**

```json
{
  "data": {
    "id": "uuid",
    "name": "My API Key",
    "key": "ak_a1b2c3d4e5f6...",
    "prefix": "ak_a1b2c3...",
    "createdAt": "2025-01-01T00:00:00Z"
  }
}
```

| フィールド | 型 | 説明 |
|-----------|-----|------|
| id | UUID | API キー ID |
| name | String | 管理名 |
| key | String | API キー平文（この応答でのみ返却） |
| prefix | String | 表示用プレフィックス |
| createdAt | DateTime | 作成日時 |

### エラー

| ステータス | コード | 説明 |
|-----------|--------|------|
| 400 | VALIDATION_ERROR | バリデーションエラー |
| 401 | UNAUTHORIZED | 未認証 |
| 409 | DUPLICATE_NAME | 同名の API キーが既に存在する |

---

## API キー一覧取得

自分の API キー一覧を取得する。平文キーは含まれない。

### リクエスト

```
GET /api/v1/me/api-keys
Authorization: Bearer {token}
```

### レスポンス

**200 OK**

```json
{
  "data": [
    {
      "id": "uuid",
      "name": "My API Key",
      "prefix": "ak_a1b2c3...",
      "lastUsedAt": "2025-01-01T00:00:00Z",
      "createdAt": "2025-01-01T00:00:00Z"
    }
  ]
}
```

| フィールド | 型 | 説明 |
|-----------|-----|------|
| id | UUID | API キー ID |
| name | String | 管理名 |
| prefix | String | 表示用プレフィックス |
| lastUsedAt | DateTime | 最終使用日時（未使用の場合は null） |
| createdAt | DateTime | 作成日時 |

### エラー

| ステータス | コード | 説明 |
|-----------|--------|------|
| 401 | UNAUTHORIZED | 未認証 |

---

## API キー削除

指定した API キーを削除する。

### リクエスト

```
DELETE /api/v1/me/api-keys/:apiKeyId
Authorization: Bearer {token}
```

### レスポンス

**204 No Content**

### エラー

| ステータス | コード | 説明 |
|-----------|--------|------|
| 401 | UNAUTHORIZED | 未認証 |
| 404 | NOT_FOUND | API キーが見つからない |
