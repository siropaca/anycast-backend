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
| categorySlug | string | - | カテゴリスラッグでフィルタ |
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

# Reactions（リアクション）

エピソードへのリアクション機能（like / bad）。同じエピソードには 1 つのリアクションのみ設定可能（排他的）。

## リアクション登録・更新

```
POST /episodes/:episodeId/reactions
```

リアクションを登録する。既に同じエピソードにリアクションがある場合は更新する（upsert）。

**リクエスト:**
```json
{
  "reactionType": "like"
}
```

| フィールド | 型 | 必須 | 説明 |
|------------|-----|:----:|------|
| reactionType | string | ◯ | リアクションタイプ（`like` / `bad`） |

**レスポンス（201 Created / 200 OK）:**

- 201: 新規登録時
- 200: 既存リアクションの更新時

```json
{
  "data": {
    "id": "uuid",
    "episodeId": "uuid",
    "reactionType": "like",
    "createdAt": "2025-01-01T00:00:00Z"
  }
}

---

## リアクション解除

```
DELETE /episodes/:episodeId/reactions
```

**レスポンス（204 No Content）:**
レスポンスボディなし

**エラー（404 Not Found）:**
```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "リアクションが見つかりません"
  }
}
```

---

## 高評価したエピソード一覧

like したエピソードの一覧を取得する。

```
GET /me/likes
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

# Playlists（プレイリスト）

YouTube 式のプレイリスト機能。各ユーザーにはデフォルトプレイリスト（後で聴く）が自動作成される。

## プレイリスト一覧取得

```
GET /me/playlists
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
      "id": "uuid",
      "name": "後で聴く",
      "description": "",
      "isDefault": true,
      "itemCount": 5,
      "createdAt": "2025-01-01T00:00:00Z",
      "updatedAt": "2025-01-01T00:00:00Z"
    }
  ],
  "pagination": {
    "total": 3,
    "limit": 20,
    "offset": 0
  }
}
```

---

## プレイリスト詳細取得

```
GET /me/playlists/:playlistId
```

**レスポンス:**
```json
{
  "data": {
    "id": "uuid",
    "name": "お気に入り",
    "description": "お気に入りのエピソード",
    "isDefault": false,
    "items": [
      {
        "id": "uuid",
        "position": 0,
        "episode": {
          "id": "uuid",
          "title": "エピソードタイトル",
          "description": "説明",
          "artwork": { "id": "uuid", "url": "..." },
          "fullAudio": { "id": "uuid", "url": "...", "durationMs": 180000 },
          "playCount": 100,
          "publishedAt": "2025-01-01T00:00:00Z",
          "channel": {
            "id": "uuid",
            "name": "チャンネル名",
            "artwork": { "id": "uuid", "url": "..." }
          }
        },
        "addedAt": "2025-01-01T00:00:00Z"
      }
    ],
    "createdAt": "2025-01-01T00:00:00Z",
    "updatedAt": "2025-01-01T00:00:00Z"
  }
}
```

---

## プレイリスト作成

```
POST /me/playlists
```

**リクエスト:**
```json
{
  "name": "お気に入り",
  "description": "お気に入りのエピソード"
}
```

| フィールド | 型 | 必須 | 説明 |
|------------|-----|:----:|------|
| name | string | ◯ | プレイリスト名（100文字以内） |
| description | string | | 説明（500文字以内） |

**レスポンス（201 Created）:**
```json
{
  "data": {
    "id": "uuid",
    "name": "お気に入り",
    "description": "お気に入りのエピソード",
    "isDefault": false,
    "itemCount": 0,
    "createdAt": "2025-01-01T00:00:00Z",
    "updatedAt": "2025-01-01T00:00:00Z"
  }
}
```

**エラー（409 Conflict）:**
```json
{
  "error": {
    "code": "DUPLICATE_NAME",
    "message": "この名前のプレイリストは既に存在します"
  }
}
```

---

## プレイリスト更新

```
PATCH /me/playlists/:playlistId
```

**リクエスト:**
```json
{
  "name": "新しい名前",
  "description": "新しい説明"
}
```

| フィールド | 型 | 必須 | 説明 |
|------------|-----|:----:|------|
| name | string | | プレイリスト名（100文字以内、デフォルトプレイリストは変更不可） |
| description | string | | 説明（500文字以内） |

**レスポンス（200 OK）:**
```json
{
  "data": {
    "id": "uuid",
    "name": "新しい名前",
    "description": "新しい説明",
    "isDefault": false,
    "itemCount": 5,
    "createdAt": "2025-01-01T00:00:00Z",
    "updatedAt": "2025-01-01T00:00:01Z"
  }
}
```

**エラー（409 Conflict）:**
```json
{
  "error": {
    "code": "DEFAULT_PLAYLIST",
    "message": "デフォルトプレイリストの名前は変更できません"
  }
}
```

---

## プレイリスト削除

```
DELETE /me/playlists/:playlistId
```

デフォルトプレイリストは削除不可。

**レスポンス（204 No Content）:**
レスポンスボディなし

**エラー（409 Conflict）:**
```json
{
  "error": {
    "code": "DEFAULT_PLAYLIST",
    "message": "デフォルトプレイリストは削除できません"
  }
}
```

---

## プレイリストにアイテム追加

```
POST /me/playlists/:playlistId/items
```

**リクエスト:**
```json
{
  "episodeId": "uuid"
}
```

| フィールド | 型 | 必須 | 説明 |
|------------|-----|:----:|------|
| episodeId | uuid | ◯ | 追加するエピソードの ID |

**レスポンス（201 Created）:**
```json
{
  "data": {
    "id": "uuid",
    "position": 0,
    "episode": {
      "id": "uuid",
      "title": "エピソードタイトル",
      "description": "説明",
      "artwork": { "id": "uuid", "url": "..." },
      "fullAudio": { "id": "uuid", "url": "...", "durationMs": 180000 },
      "playCount": 100,
      "publishedAt": "2025-01-01T00:00:00Z",
      "channel": {
        "id": "uuid",
        "name": "チャンネル名",
        "artwork": { "id": "uuid", "url": "..." }
      }
    },
    "addedAt": "2025-01-01T00:00:00Z"
  }
}
```

**エラー（409 Conflict）:**
```json
{
  "error": {
    "code": "ALREADY_IN_PLAYLIST",
    "message": "このエピソードは既にプレイリストに追加されています"
  }
}
```

---

## プレイリストからアイテム削除

```
DELETE /me/playlists/:playlistId/items/:itemId
```

**レスポンス（204 No Content）:**
レスポンスボディなし

---

## プレイリストアイテム並び替え

```
POST /me/playlists/:playlistId/items/reorder
```

**リクエスト:**
```json
{
  "itemIds": ["uuid1", "uuid2", "uuid3"]
}
```

| フィールド | 型 | 必須 | 説明 |
|------------|-----|:----:|------|
| itemIds | uuid[] | ◯ | 新しい順序でのアイテム ID 配列 |

**レスポンス（200 OK）:**
プレイリスト詳細と同じ形式

---

## デフォルトプレイリスト（後で聴く）一覧取得

デフォルトプレイリスト（後で聴く）の内容を取得するショートカット。

```
GET /me/default-playlist
```

**レスポンス:**
プレイリスト詳細と同じ形式

---

## デフォルトプレイリスト（後で聴く）に追加

```
POST /episodes/:episodeId/default-playlist
```

**レスポンス（201 Created）:**
プレイリストアイテムと同じ形式

**エラー（409 Conflict）:**
```json
{
  "error": {
    "code": "ALREADY_IN_PLAYLIST",
    "message": "このエピソードは既にプレイリストに追加されています"
  }
}
```

---

## デフォルトプレイリスト（後で聴く）から削除

```
DELETE /episodes/:episodeId/default-playlist
```

**レスポンス（204 No Content）:**
レスポンスボディなし

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
GET /me/playback-history
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

他のユーザーをフォローする機能。自分自身はフォロー不可。

## フォロー状態取得

```
GET /users/:username/follow
```

指定ユーザーをフォローしているかどうかを返す。

**レスポンス（200 OK）:**
```json
{
  "data": {
    "following": true
  }
}
```

**エラー（404 Not Found）:**
```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "ユーザーが見つかりません"
  }
}
```

---

## フォロー登録

```
POST /users/:username/follow
```

**レスポンス（201 Created）:**
```json
{
  "data": {
    "id": "uuid",
    "targetUserId": "uuid",
    "createdAt": "2025-01-01T00:00:00Z"
  }
}
```

**エラー（400 Bad Request）:**
```json
{
  "error": {
    "code": "SELF_FOLLOW_NOT_ALLOWED",
    "message": "自分自身はフォローできません"
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
DELETE /users/:username/follow
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

## フォロー中のユーザー一覧

```
GET /me/follows
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
      "user": {
        "id": "uuid",
        "username": "user_name",
        "displayName": "ユーザー名",
        "avatar": { "id": "uuid", "url": "..." }
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

---

# Comments（コメント）

エピソードへのコメント機能。

## コメント投稿

```
POST /episodes/:episodeId/comments
```

**リクエスト:**
```json
{
  "content": "とても参考になりました！"
}
```

| フィールド | 型 | 必須 | 説明 |
|------------|-----|:----:|------|
| content | string | ◯ | コメント本文（1〜1000文字） |

**レスポンス（201 Created）:**
```json
{
  "data": {
    "id": "uuid",
    "user": {
      "id": "uuid",
      "username": "user_name",
      "displayName": "ユーザー名",
      "avatar": { "id": "uuid", "url": "..." }
    },
    "episodeId": "uuid",
    "content": "とても参考になりました！",
    "createdAt": "2025-01-01T00:00:00Z",
    "updatedAt": "2025-01-01T00:00:00Z"
  }
}
```

**エラー（400 Bad Request）:**
```json
{
  "error": {
    "code": "INVALID_CONTENT_LENGTH",
    "message": "コメントは1〜1000文字で入力してください"
  }
}
```

---

## コメント一覧取得

```
GET /episodes/:episodeId/comments
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
      "id": "uuid",
      "user": {
        "id": "uuid",
        "username": "user_name",
        "displayName": "ユーザー名",
        "avatar": { "id": "uuid", "url": "..." }
      },
      "content": "とても参考になりました！",
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

※ 削除されたコメント（deleted_at が設定されているもの）は表示されない

---

## コメント編集

```
PATCH /comments/:commentId
```

コメント投稿者本人または Admin のみ編集可能。

**リクエスト:**
```json
{
  "content": "修正しました。とても参考になりました！"
}
```

| フィールド | 型 | 必須 | 説明 |
|------------|-----|:----:|------|
| content | string | ◯ | コメント本文（1〜1000文字） |

**レスポンス（200 OK）:**
```json
{
  "data": {
    "id": "uuid",
    "user": {
      "id": "uuid",
      "username": "user_name",
      "displayName": "ユーザー名",
      "avatar": { "id": "uuid", "url": "..." }
    },
    "episodeId": "uuid",
    "content": "修正しました。とても参考になりました！",
    "createdAt": "2025-01-01T00:00:00Z",
    "updatedAt": "2025-01-01T00:00:01Z"
  }
}
```

**エラー（403 Forbidden）:**
```json
{
  "error": {
    "code": "FORBIDDEN",
    "message": "このコメントを編集する権限がありません"
  }
}
```

---

## コメント削除

```
DELETE /comments/:commentId
```

コメント投稿者本人または Admin のみ削除可能。論理削除。

**レスポンス（204 No Content）:**
レスポンスボディなし

**エラー（403 Forbidden）:**
```json
{
  "error": {
    "code": "FORBIDDEN",
    "message": "このコメントを削除する権限がありません"
  }
}
```

**エラー（404 Not Found）:**
```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "コメントが見つかりません"
  }
}
```

---

## 自分のコメント一覧

```
GET /me/comments
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
      "id": "uuid",
      "episode": {
        "id": "uuid",
        "title": "エピソードタイトル",
        "channel": {
          "id": "uuid",
          "name": "チャンネル名"
        }
      },
      "content": "とても参考になりました！",
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
