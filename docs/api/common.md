# API 共通仕様

## レスポンス形式

**成功時:**
```json
{
  "data": { ... }
}
```

**エラー時:**
```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "エラーメッセージ"
  }
}
```

## ページネーション

一覧取得 API は以下のクエリパラメータをサポート:

| パラメータ | 型 | デフォルト | 説明 |
|------------|-----|------------|------|
| limit | int | 20 | 取得件数（最大 100） |
| offset | int | 0 | オフセット |

**レスポンス:**
```json
{
  "data": [ ... ],
  "pagination": {
    "total": 100,
    "limit": 20,
    "offset": 0
  }
}
```

## 権限

| 権限レベル | 説明 |
|------------|------|
| Guest | ログインなしでアクセス可能（現時点では該当なし） |
| Public | ログイン済みユーザーであれば誰でもアクセス可能 |
| Owner | 自身のリソースのみ操作可能 |
| Admin | 運営のみ操作可能 |

**エンドポイント別権限:**

| リソース | 参照 | 作成 | 更新 | 削除 |
|----------|:----:|:----:|:----:|:----:|
| Users | Owner | - | Owner | Owner |
| Channels | Public | Owner | Owner | Owner |
| Characters | Owner | Owner | Owner | Owner |
| BGMs | Owner | Owner | Owner | Owner |
| Default BGMs | Public | Admin | Admin | Admin |
| Episodes | Public | Owner | Owner | Owner |
| Script / ScriptLines | Owner | Owner | Owner | Owner |
| Likes | Owner | Owner | - | Owner |
| Bookmarks | Owner | Owner | - | Owner |
| Playback History | Owner | Owner | Owner | Owner |
| Follows | Owner | Owner | - | Owner |
| Audio（生成） | - | Owner | - | - |
| Audios（アップロード） | Owner | Owner | - | Owner |
| Images（アップロード） | Owner | Owner | - | Owner |
| Voices | Public | Admin | Admin | Admin |
| Categories | Public | Admin | Admin | Admin |

## 公開状態によるアクセス制御

チャンネルとエピソードには公開状態（`publishedAt`）があり、参照時のアクセス制御に影響する。

| エンドポイント | オーナー | 他ユーザー |
|---------------|----------|------------|
| `GET /channels` | - | 公開中のみ |
| `GET /channels/:channelId` | 全て | 公開中のみ |
| `GET /channels/:channelId/episodes` | - | 公開中のみ |
| `GET /channels/:channelId/episodes/:episodeId` | 全て | 公開中のみ |
| `GET /search/channels` | - | 公開中のみ |
| `GET /search/episodes` | - | 公開中のみ |
| `GET /me/channels` | 全て | - |
| `GET /me/channels/:channelId/episodes` | 全て | - |

- **公開中**: `publishedAt IS NOT NULL AND publishedAt <= NOW()`
- **非公開（下書き）**: `publishedAt IS NULL`
- **予約公開**: `publishedAt > NOW()`（将来的に対応可能）
