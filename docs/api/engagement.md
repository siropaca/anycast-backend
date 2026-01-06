# Search（検索）

フリーワード検索用のエンドポイント。公開中のコンテンツのみ検索可能。

## チャンネル検索

```
GET /search/channels
```

**クエリパラメータ:**

| パラメータ | 型 | デフォルト | 説明 |
|------------|-----|------------|------|
| q | string | **必須** | 検索キーワード（name, description を対象） |
| categoryId | uuid | - | カテゴリ ID でフィルタ |
| limit | int | 20 | 取得件数（最大 100） |
| offset | int | 0 | オフセット |

**レスポンス:**
```json
{
  "data": [
    {
      "id": "uuid",
      "name": "チャンネル名",
      "description": "説明",
      "category": { "id": "uuid", "slug": "technology", "name": "テクノロジー" },
      "artwork": { "id": "uuid", "url": "..." },
      "publishedAt": "2025-01-01T00:00:00Z",
      "createdAt": "2025-01-01T00:00:00Z",
      "updatedAt": "2025-01-01T00:00:00Z"
    }
  ],
  "pagination": {
    "total": 100,
    "limit": 20,
    "offset": 0
  }
}
```

---

## エピソード検索

```
GET /search/episodes
```

**クエリパラメータ:**

| パラメータ | 型 | デフォルト | 説明 |
|------------|-----|------------|------|
| q | string | **必須** | 検索キーワード（title, description を対象） |
| limit | int | 20 | 取得件数（最大 100） |
| offset | int | 0 | オフセット |

**レスポンス:**
```json
{
  "data": [
    {
      "id": "uuid",
      "title": "エピソードタイトル",
      "description": "エピソードの説明",
      "channel": {
        "id": "uuid",
        "name": "チャンネル名"
      },
      "publishedAt": "2025-01-01T00:00:00Z",
      "createdAt": "2025-01-01T00:00:00Z",
      "updatedAt": "2025-01-01T00:00:00Z"
    }
  ],
  "pagination": {
    "total": 100,
    "limit": 20,
    "offset": 0
  }
}
```

---

## ユーザー検索

```
GET /search/users
```

**クエリパラメータ:**

| パラメータ | 型 | デフォルト | 説明 |
|------------|-----|------------|------|
| q | string | **必須** | 検索キーワード（username, displayName を対象） |
| limit | int | 20 | 取得件数（最大 100） |
| offset | int | 0 | オフセット |

**レスポンス:**
```json
{
  "data": [
    {
      "id": "uuid",
      "username": "user_name",
      "displayName": "ユーザー名",
      "avatar": { "id": "uuid", "url": "..." },
      "createdAt": "2025-01-01T00:00:00Z"
    }
  ],
  "pagination": {
    "total": 100,
    "limit": 20,
    "offset": 0
  }
}
```

※ 検索結果は公開プロフィールのみ（email は非公開）

---

# Likes（お気に入り）

エピソードへのお気に入り機能。

## お気に入り登録

```
POST /episodes/:episodeId/like
```

**レスポンス（201 Created）:**
```json
{
  "data": {
    "id": "uuid",
    "episodeId": "uuid",
    "createdAt": "2025-01-01T00:00:00Z"
  }
}
```

**エラー（409 Conflict）:**
```json
{
  "error": {
    "code": "ALREADY_LIKED",
    "message": "既にお気に入り済みです"
  }
}
```

---

## お気に入り解除

```
DELETE /episodes/:episodeId/like
```

**レスポンス（204 No Content）:**
レスポンスボディなし

**エラー（404 Not Found）:**
```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "お気に入りが見つかりません"
  }
}
```

---

## お気に入りしたエピソード一覧

```
GET /auth/me/likes
```

**クエリパラメータ:**

| パラメータ | 型 | デフォルト | 説明 |
|------------|-----|------------|------|
| limit | int | 20 | 取得件数（最大 100） |
| offset | int | 0 | オフセット |

**レスポンス:**
```json
{
  "data": [
    {
      "episode": {
        "id": "uuid",
        "title": "エピソードタイトル",
        "description": "説明",
        "channel": {
          "id": "uuid",
          "name": "チャンネル名",
          "artwork": { "id": "uuid", "url": "..." }
        },
        "publishedAt": "2025-01-01T00:00:00Z"
      },
      "likedAt": "2025-01-01T00:00:00Z"
    }
  ],
  "pagination": {
    "total": 100,
    "limit": 20,
    "offset": 0
  }
}
```

---

# Bookmarks（後で見る）

エピソードへの「後で見る」機能。

## ブックマーク登録

```
POST /episodes/:episodeId/bookmark
```

**レスポンス（201 Created）:**
```json
{
  "data": {
    "id": "uuid",
    "episodeId": "uuid",
    "createdAt": "2025-01-01T00:00:00Z"
  }
}
```

**エラー（409 Conflict）:**
```json
{
  "error": {
    "code": "ALREADY_BOOKMARKED",
    "message": "既にブックマーク済みです"
  }
}
```

---

## ブックマーク解除

```
DELETE /episodes/:episodeId/bookmark
```

**レスポンス（204 No Content）:**
レスポンスボディなし

**エラー（404 Not Found）:**
```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "ブックマークが見つかりません"
  }
}
```

---

## ブックマークしたエピソード一覧

```
GET /auth/me/bookmarks
```

**クエリパラメータ:**

| パラメータ | 型 | デフォルト | 説明 |
|------------|-----|------------|------|
| limit | int | 20 | 取得件数（最大 100） |
| offset | int | 0 | オフセット |

**レスポンス:**
```json
{
  "data": [
    {
      "episode": {
        "id": "uuid",
        "title": "エピソードタイトル",
        "description": "説明",
        "channel": {
          "id": "uuid",
          "name": "チャンネル名",
          "artwork": { "id": "uuid", "url": "..." }
        },
        "publishedAt": "2025-01-01T00:00:00Z"
      },
      "bookmarkedAt": "2025-01-01T00:00:00Z"
    }
  ],
  "pagination": {
    "total": 100,
    "limit": 20,
    "offset": 0
  }
}
```

---

# Playback History（再生履歴）

エピソードの再生履歴を管理する。

## 再生履歴を更新

```
PUT /episodes/:episodeId/playback
```

再生位置や完了状態を更新する。履歴が存在しない場合は新規作成（Upsert）。

**リクエスト:**
```json
{
  "progressMs": 120000,
  "completed": false
}
```

| フィールド | 型 | 必須 | 説明 |
|------------|-----|:----:|------|
| progressMs | int | | 再生位置（ミリ秒） |
| completed | bool | | 再生完了フラグ |

**レスポンス（200 OK）:**
```json
{
  "data": {
    "episodeId": "uuid",
    "progressMs": 120000,
    "completed": false,
    "playedAt": "2025-01-01T00:00:00Z"
  }
}
```

---

## 再生履歴を削除

```
DELETE /episodes/:episodeId/playback
```

**レスポンス（204 No Content）:**
レスポンスボディなし

**エラー（404 Not Found）:**
```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "再生履歴が見つかりません"
  }
}
```

---

## 再生履歴一覧を取得

```
GET /auth/me/playback-history
```

最近再生した順で一覧を取得。

**クエリパラメータ:**

| パラメータ | 型 | デフォルト | 説明 |
|------------|-----|------------|------|
| completed | bool | - | 完了状態でフィルタ |
| limit | int | 20 | 取得件数（最大 100） |
| offset | int | 0 | オフセット |

**レスポンス:**
```json
{
  "data": [
    {
      "episode": {
        "id": "uuid",
        "title": "エピソードタイトル",
        "description": "説明",
        "fullAudio": { "id": "uuid", "url": "...", "durationMs": 180000 },
        "channel": {
          "id": "uuid",
          "name": "チャンネル名",
          "artwork": { "id": "uuid", "url": "..." }
        },
        "publishedAt": "2025-01-01T00:00:00Z"
      },
      "progressMs": 120000,
      "completed": false,
      "playedAt": "2025-01-01T00:00:00Z"
    }
  ],
  "pagination": {
    "total": 100,
    "limit": 20,
    "offset": 0
  }
}
```

---

# Follows（フォロー）

他のユーザーのエピソードをフォローする機能。自分のチャンネルのエピソードはフォロー不可。

## フォロー登録

```
POST /episodes/:episodeId/follow
```

**レスポンス（201 Created）:**
```json
{
  "data": {
    "id": "uuid",
    "episodeId": "uuid",
    "createdAt": "2025-01-01T00:00:00Z"
  }
}
```

**エラー（400 Bad Request）:**
```json
{
  "error": {
    "code": "SELF_FOLLOW_NOT_ALLOWED",
    "message": "自分のエピソードはフォローできません"
  }
}
```

**エラー（409 Conflict）:**
```json
{
  "error": {
    "code": "ALREADY_FOLLOWED",
    "message": "既にフォロー済みです"
  }
}
```

---

## フォロー解除

```
DELETE /episodes/:episodeId/follow
```

**レスポンス（204 No Content）:**
レスポンスボディなし

**エラー（404 Not Found）:**
```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "フォローが見つかりません"
  }
}
```

---

## フォロー中のエピソード一覧

```
GET /auth/me/follows
```

**クエリパラメータ:**

| パラメータ | 型 | デフォルト | 説明 |
|------------|-----|------------|------|
| limit | int | 20 | 取得件数（最大 100） |
| offset | int | 0 | オフセット |

**レスポンス:**
```json
{
  "data": [
    {
      "episode": {
        "id": "uuid",
        "title": "エピソードタイトル",
        "description": "説明",
        "channel": {
          "id": "uuid",
          "name": "チャンネル名",
          "artwork": { "id": "uuid", "url": "..." }
        },
        "publishedAt": "2025-01-01T00:00:00Z"
      },
      "followedAt": "2025-01-01T00:00:00Z"
    }
  ],
  "pagination": {
    "total": 100,
    "limit": 20,
    "offset": 0
  }
}
```
