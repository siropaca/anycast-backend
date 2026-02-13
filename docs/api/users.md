# Users（ユーザー）

## ユーザー取得

```
GET /users/:username
```

**レスポンス:**
```json
{
  "data": {
    "id": "uuid",
    "username": "user_name",
    "displayName": "ユーザー名",
    "bio": "自己紹介文",
    "avatar": { "id": "uuid", "url": "..." },
    "headerImage": { "id": "uuid", "url": "..." },
    "followingCount": 10,
    "followerCount": 25,
    "channels": [
      {
        "id": "uuid",
        "name": "チャンネル名",
        "description": "説明",
        "category": { "id": "uuid", "slug": "technology", "name": "テクノロジー" },
        "artwork": { "id": "uuid", "url": "..." },
        "episodeCount": 12,
        "publishedAt": "2025-01-01T00:00:00Z",
        "createdAt": "2025-01-01T00:00:00Z",
        "updatedAt": "2025-01-01T00:00:00Z"
      }
    ],
    "createdAt": "2025-01-01T00:00:00Z"
  }
}
```

> **Note:**
> - 他ユーザーの情報は公開プロフィールのみ（email は非公開）
> - `channels` には公開中のチャンネルのみ含まれます
> - `episodeCount` は公開済みエピソードの件数です

---

## 現在のユーザー取得

```
GET /me
```

**レスポンス:**
```json
{
  "data": {
    "id": "uuid",
    "email": "user@example.com",
    "username": "user_name",
    "displayName": "ユーザー名",
    "bio": "自己紹介文",
    "avatar": { "id": "uuid", "url": "..." },
    "headerImage": { "id": "uuid", "url": "..." },
    "userPrompt": "台本生成の基本方針...",
    "hasPassword": true,
    "oauthProviders": ["google"],
    "createdAt": "2025-01-01T00:00:00Z"
  }
}
```

---

## ユーザー情報更新

```
PATCH /me
```

プロフィール情報（表示名、自己紹介、アバター画像、ヘッダー画像）を更新する。

**リクエスト:**
```json
{
  "displayName": "新しい名前",
  "bio": "自己紹介文",
  "avatarImageId": "uuid",
  "headerImageId": "uuid"
}
```

| フィールド | 型 | 必須 | 説明 |
|-----------|-----|:----:|------|
| displayName | String | ◯ | 表示名（20文字以内） |
| bio | String | | 自己紹介文（200文字以内、空文字可） |
| avatarImageId | String? | | アバター画像 ID。`null`=変更なし、`""`=クリア、UUID=設定 |
| headerImageId | String? | | ヘッダー画像 ID。`null`=変更なし、`""`=クリア、UUID=設定 |

**レスポンス:** `GET /me` と同じ形式

---

## ユーザー名変更

```
PATCH /me/username
```

ユーザー名（username）を変更する。一意性チェックを行い、重複する場合は 409 エラーを返す。

**リクエスト:**
```json
{
  "username": "new_username"
}
```

| フィールド | 型 | 必須 | 説明 |
|-----------|-----|:----:|------|
| username | String | ◯ | 新しいユーザー名（3〜20文字、英数字・アンダースコア・日本語、`__` 始まり禁止） |

**バリデーションルール:**
- 3〜20文字
- 使用可能文字: 英数字、アンダースコア（`_`）、日本語（ひらがな・カタカナ・漢字）
- `__`（アンダースコア2つ）で始まるユーザー名は禁止（システム予約）
- 大文字小文字を区別する
- 他のユーザーと重複不可

**レスポンス:** `GET /me` と同じ形式

**エラー:**
| コード | 説明 |
|--------|------|
| 400 | バリデーションエラー（形式不正、文字数超過など） |
| 409 | そのユーザー名は既に使用されています |

---

## ユーザー名利用可否チェック

```
GET /me/username/check?username=new_username
```

指定したユーザー名が利用可能かどうかを確認する。ユーザー名変更画面でリアルタイムバリデーションに使用する。

**クエリパラメータ:**

| パラメータ | 型 | 必須 | 説明 |
|-----------|-----|:----:|------|
| username | String | ◯ | チェック対象のユーザー名 |

**レスポンス:**
```json
{
  "data": {
    "username": "new_username",
    "available": true
  }
}
```

| フィールド | 型 | 説明 |
|-----------|-----|------|
| username | String | チェックしたユーザー名 |
| available | Boolean | 利用可能かどうか |

> **Note:**
> - バリデーションルールは「ユーザー名変更」と同一
> - 自分の現在のユーザー名を指定した場合は `available: true` を返す
> - バリデーションエラー（形式不正）の場合は 400 エラーを返す

---

## ユーザープロンプト更新

```
PATCH /me/prompt
```

ユーザーの台本生成用プロンプト（基本方針）を更新する。このプロンプトはユーザーのすべてのチャンネル・エピソードの台本生成に適用される。

**リクエスト:**
```json
{
  "userPrompt": "台本生成の基本方針..."
}
```

**レスポンス:**
```json
{
  "data": {
    "id": "uuid",
    "email": "user@example.com",
    "username": "user_name",
    "displayName": "ユーザー名",
    "bio": "自己紹介文",
    "avatar": { "id": "uuid", "url": "..." },
    "headerImage": { "id": "uuid", "url": "..." },
    "userPrompt": "台本生成の基本方針...",
    "hasPassword": true,
    "oauthProviders": ["google"],
    "createdAt": "2025-01-01T00:00:00Z"
  }
}
```

### プロンプトの適用順序

台本生成時、プロンプトは以下の順序で結合（追記）される：

1. **User.userPrompt** - ユーザーの基本方針
2. **Channel.userPrompt** - チャンネル固有の方針

---

## アカウント削除

```
DELETE /me
```

自分のアカウントを削除する。アカウントに紐づくすべてのデータが物理削除される。実行中のジョブがある場合はキャンセルしてから削除する。

**処理の流れ:**

1. 実行中の音声生成ジョブ・台本生成ジョブをキャンセル
2. ユーザーレコードを物理削除（関連データは ON DELETE CASCADE で連鎖削除）

**削除されるデータ:**

| データ | 削除方法 |
|--------|----------|
| credentials | CASCADE |
| oauth_accounts | CASCADE |
| refresh_tokens | CASCADE |
| channels（→ episodes → script_lines, audio_jobs） | CASCADE |
| characters | CASCADE |
| bgms | CASCADE |
| playlists（→ playlist_items） | CASCADE |
| reactions | CASCADE |
| playback_histories | CASCADE |
| follows（フォロー・被フォロー両方） | CASCADE |
| comments | CASCADE |
| feedbacks | CASCADE |

**削除されないデータ:**

| データ | 挙動 |
|--------|------|
| contacts | `user_id` が SET NULL（お問い合わせ履歴は保持） |
| images / audios | 参照が外れ孤児化（バッチ処理でクリーンアップ） |

**レスポンス:** `204 No Content`

**エラー:**

| コード | 説明 |
|--------|------|
| 401 | 認証が必要 |
