# Episodes（エピソード）

## エピソード一覧取得（公開用）

```
GET /channels/:channelId/episodes
```

認証不要。公開中のエピソードは誰でも取得可能。認証済みの場合、自分のチャンネルのエピソードは非公開でも取得可能。非公開含む全エピソードの管理は `GET /me/channels/:channelId/episodes` を使用。

**クエリパラメータ:**

| パラメータ | 型 | デフォルト | 説明 |
|------------|-----|------------|------|
| sort | string | `createdAt` | ソート対象: `createdAt` / `updatedAt` |
| order | string | `asc` | ソート順: `asc`（昇順） / `desc`（降順） |
| limit | int | 20 | 取得件数（最大 100） |
| offset | int | 0 | オフセット |

**レスポンス:**
```json
{
  "data": [
    {
      "id": "uuid",
      "owner": { "id": "uuid", "username": "testuser", "displayName": "テストユーザー", "avatar": null },
      "title": "エピソードタイトル",
      "description": "エピソードの説明",
      "fullAudio": { "id": "uuid", "url": "...", "durationMs": 180000 },
      "scriptLineCount": 12,
      "playCount": 123,
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

## エピソード取得

```
GET /channels/:channelId/episodes/:episodeId
```

認証不要。公開中のエピソードは誰でも取得可能。認証済みの場合、自分のチャンネルのエピソードは非公開でも取得可能。

**レスポンス:**
```json
{
  "data": {
    "id": "uuid",
    "owner": { "id": "uuid", "username": "testuser", "displayName": "テストユーザー", "avatar": null },
    "title": "エピソードタイトル",
    "description": "エピソードの説明",
    "bgm": { "id": "uuid", "url": "..." },
    "voiceAudio": { "id": "uuid", "url": "...", "durationMs": 120000 },
    "fullAudio": { "id": "uuid", "url": "..." },
    "script": [
      {
        "id": "uuid",
        "lineOrder": 0,
        "speaker": { "id": "uuid", "name": "太郎" },
        "text": "こんにちは",
        "emotion": null
      },
      {
        "id": "uuid",
        "lineOrder": 1,
        "speaker": { "id": "uuid", "name": "花子" },
        "text": "やあ、元気？",
        "emotion": "嬉しそうに"
      }
    ],
    "playback": {
      "progressMs": 60000,
      "completed": false,
      "playedAt": "2025-01-01T00:00:00Z"
    },
    "playlistIds": ["uuid1", "uuid2"],
    "playCount": 123,
    "publishedAt": "2025-01-01T00:00:00Z",
    "createdAt": "2025-01-01T00:00:00Z",
    "updatedAt": "2025-01-01T00:00:00Z"
  }
}
```

> **Note:** `voiceAudio` はオーナーのみに表示されます。他ユーザーがアクセスした場合は含まれません。
>
> **Note:** `playback` は認証済みの場合のみ含まれます。未認証または再生履歴がない場合は `null` になります。
>
> **Note:** `playlistIds` は認証済みの場合のみ含まれます。未認証の場合は `null` になります。エピソードがどの再生リストにも含まれていない場合は空配列 `[]` になります。
>
> **Note:** BGM の設定・削除は専用エンドポイント（[エピソード BGM 設定](#エピソード-bgm-設定) / [エピソード BGM 削除](#エピソード-bgm-削除)）を使用してください。

---

## エピソード作成

```
POST /channels/:channelId/episodes
```

**リクエスト:**
```json
{
  "title": "エピソードタイトル",
  "description": "エピソードの説明",
  "artworkImageId": "uuid"
}
```

**バリデーション:**
| フィールド | ルール |
|------------|--------|
| title | 必須、255文字以内 |
| description | 必須、2000文字以内 |

---

## エピソード更新

```
PATCH /channels/:channelId/episodes/:episodeId
```

**リクエスト:**
```json
{
  "title": "新しいタイトル",
  "description": "新しい説明",
  "artworkImageId": "uuid"
}
```

**バリデーション:**
| フィールド | ルール |
|------------|--------|
| title | 255文字以内 |
| description | 2000文字以内 |

> **Note:** 公開状態の変更は専用エンドポイント（[エピソード公開](#エピソード公開) / [エピソード非公開](#エピソード非公開)）を使用してください。
>
> **Note:** BGM の変更は専用エンドポイント（[エピソード BGM 設定](#エピソード-bgm-設定) / [エピソード BGM 削除](#エピソード-bgm-削除)）を使用してください。

---

## エピソード削除

```
DELETE /channels/:channelId/episodes/:episodeId
```

---

## エピソード公開

```
POST /channels/:channelId/episodes/:episodeId/publish
```

エピソードを公開状態にする。`publishedAt` を省略すると現在時刻で即時公開、指定すると予約公開になる。

**リクエスト:**
```json
{
  "publishedAt": "2025-01-01T00:00:00Z"
}
```

| フィールド | 型 | 必須 | 説明 |
|------------|-----|:----:|------|
| publishedAt | string | | 公開日時（RFC3339 形式）。省略時は現在時刻で即時公開 |

**レスポンス（200 OK）:**
```json
{
  "data": {
    "id": "uuid",
    "title": "エピソードタイトル",
    "description": "エピソードの説明",
    "publishedAt": "2025-01-01T00:00:00Z",
    "createdAt": "2025-01-01T00:00:00Z",
    "updatedAt": "2025-01-01T00:00:00Z"
  }
}
```

---

## エピソード非公開

```
POST /channels/:channelId/episodes/:episodeId/unpublish
```

エピソードを非公開（下書き）状態に戻す。

**レスポンス（200 OK）:**
```json
{
  "data": {
    "id": "uuid",
    "title": "エピソードタイトル",
    "description": "エピソードの説明",
    "publishedAt": null,
    "createdAt": "2025-01-01T00:00:00Z",
    "updatedAt": "2025-01-01T00:00:00Z"
  }
}
```

---

## エピソード BGM 設定

```
PUT /channels/:channelId/episodes/:episodeId/bgm
```

エピソードに BGM を設定する。ユーザー BGM またはシステム BGM のどちらかを指定する。

**リクエスト:**
```json
{
  "bgmId": "uuid",
  "systemBgmId": "uuid"
}
```

**バリデーション:**

| フィールド | ルール |
|------------|--------|
| bgmId | UUID 形式、自分が所有する BGM のみ |
| systemBgmId | UUID 形式、is_active = true のシステム BGM のみ |

> **Note:** `bgmId` と `systemBgmId` は同時に指定できません。いずれか一方を必ず指定してください。

**レスポンス（200 OK）:**
```json
{
  "data": {
    "id": "uuid",
    "title": "エピソードタイトル",
    "description": "エピソードの説明",
    "bgm": {
      "id": "uuid",
      "name": "Chill BGM",
      "isSystem": false,
      "audio": {
        "id": "uuid",
        "url": "https://storage.example.com/audios/xxx.mp3?signature=...",
        "durationMs": 180000
      }
    },
    "createdAt": "2025-01-01T00:00:00Z",
    "updatedAt": "2025-01-01T00:00:00Z"
  }
}
```

**エラー（400 Bad Request）:**
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "bgmId と systemBgmId は同時に指定できません"
  }
}
```

**エラー（403 Forbidden）:**
```json
{
  "error": {
    "code": "FORBIDDEN",
    "message": "このエピソードの BGM 設定権限がありません"
  }
}
```

---

## エピソード BGM 削除

```
DELETE /channels/:channelId/episodes/:episodeId/bgm
```

エピソードの BGM 設定を削除する。

**レスポンス（200 OK）:**
```json
{
  "data": {
    "id": "uuid",
    "title": "エピソードタイトル",
    "description": "エピソードの説明",
    "bgm": null,
    "createdAt": "2025-01-01T00:00:00Z",
    "updatedAt": "2025-01-01T00:00:00Z"
  }
}
```

**エラー（403 Forbidden）:**
```json
{
  "error": {
    "code": "FORBIDDEN",
    "message": "このエピソードの BGM 削除権限がありません"
  }
}
```

---

## 自分のチャンネルのエピソード一覧取得

```
GET /me/channels/:channelId/episodes
```

自分のチャンネルに紐付くエピソード一覧を取得（非公開含む）。編集画面での使用を想定。

**パスパラメータ:**

| パラメータ | 型 | 説明 |
|------------|-----|------|
| channelId | uuid | チャンネル ID |

**クエリパラメータ:**

| パラメータ | 型 | デフォルト | 説明 |
|------------|-----|------------|------|
| status | string | - | 公開状態でフィルタ: `published` / `draft` |
| sort | string | `createdAt` | ソート対象: `createdAt` / `updatedAt` |
| order | string | `asc` | ソート順: `asc`（昇順） / `desc`（降順） |
| limit | int | 20 | 取得件数（最大 100） |
| offset | int | 0 | オフセット |

**レスポンス:**
```json
{
  "data": [
    {
      "id": "uuid",
      "owner": { "id": "uuid", "username": "testuser", "displayName": "テストユーザー", "avatar": null },
      "title": "エピソードタイトル",
      "description": "エピソードの説明",
      "fullAudio": { "id": "uuid", "url": "...", "durationMs": 180000 },
      "scriptLineCount": 12,
      "playCount": 123,
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

**エラー（403 Forbidden）:**
```json
{
  "error": {
    "code": "FORBIDDEN",
    "message": "このチャンネルへのアクセス権限がありません"
  }
}
```

**エラー（404 Not Found）:**
```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "チャンネルが見つかりません"
  }
}
```

---

## 自分のチャンネルのエピソード取得

```
GET /me/channels/:channelId/episodes/:episodeId
```

自分のチャンネルに紐付くエピソードを取得（非公開含む）。編集画面での使用を想定。

**パスパラメータ:**

| パラメータ | 型 | 説明 |
|------------|-----|------|
| channelId | uuid | チャンネル ID |
| episodeId | uuid | エピソード ID |

**レスポンス:**
```json
{
  "data": {
    "id": "uuid",
    "owner": { "id": "uuid", "username": "testuser", "displayName": "テストユーザー", "avatar": null },
    "title": "エピソードタイトル",
    "description": "エピソードの説明",
    "artwork": { "id": "uuid", "url": "..." },
    "fullAudio": { "id": "uuid", "url": "...", "durationMs": 180000 },
    "playback": {
      "progressMs": 60000,
      "completed": false,
      "playedAt": "2025-01-01T00:00:00Z"
    },
    "playlistIds": ["uuid1", "uuid2"],
    "playCount": 123,
    "publishedAt": "2025-01-01T00:00:00Z",
    "createdAt": "2025-01-01T00:00:00Z",
    "updatedAt": "2025-01-01T00:00:00Z"
  }
}
```

> **Note:** `playback` は認証済みの場合のみ含まれます。再生履歴がない場合は `null` になります。
>
> **Note:** `playlistIds` は認証済みユーザーの場合のみ含まれます。エピソードがどの再生リストにも含まれていない場合は空配列 `[]` になります。

**エラー（403 Forbidden）:**
```json
{
  "error": {
    "code": "FORBIDDEN",
    "message": "このチャンネルへのアクセス権限がありません"
  }
}
```

**エラー（404 Not Found）:**
```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "エピソードが見つかりません"
  }
}
```

---

## 再生回数カウント

```
POST /episodes/:episodeId/play
```

エピソードの再生回数をインクリメントする。クライアントは再生開始から 30 秒経過した時点でこの API を呼び出す。

**パスパラメータ:**

| パラメータ | 型 | 説明 |
|------------|-----|------|
| episodeId | uuid | エピソード ID |

**レスポンス（204 No Content）:**

レスポンスボディなし。

**エラー（404 Not Found）:**
```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "エピソードが見つかりません"
  }
}
```

> **Note:** 公開中のエピソードのみカウント対象。同一ユーザーによる重複カウントを許容する（毎回 +1）。

---

> **Note:** 音声生成 API の詳細仕様は `docs/specs/audio-generate-async-api.md` を参照してください。`type` パラメータで `voice`（TTS のみ）、`full`（TTS + BGM）、`remix`（BGM 差し替え）を切り替えます。
