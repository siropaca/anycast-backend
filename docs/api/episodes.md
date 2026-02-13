# Episodes（エピソード）

## エピソード一覧取得（公開用）

```
GET /channels/:channelId/episodes
```

認証不要。公開中のエピソードは誰でも取得可能。認証済みの場合、自分のチャンネルのエピソードは非公開でも取得可能。非公開含む全エピソードの管理は `GET /me/channels/:channelId/episodes` を使用。

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
      "owner": { "id": "uuid", "username": "testuser", "displayName": "テストユーザー", "avatar": null },
      "title": "エピソードタイトル",
      "description": "エピソードの説明",
      "fullAudio": { "id": "uuid", "url": "...", "durationMs": 180000 },
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
    "voiceStyle": "Read aloud in a warm, welcoming tone",
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
> **Note:** `voiceStyle` はオーナーのみに表示されます。他ユーザーがアクセスした場合は含まれません。
>
> **Note:** `playback` は認証済みの場合のみ含まれます。未認証または再生履歴がない場合は `null` になります。
>
> **Note:** `playlistIds` は認証済みの場合のみ含まれます。未認証の場合は `null` になります。エピソードがどの再生リストにも含まれていない場合は空配列 `[]` になります。

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

> **Note:** `voiceStyle` は音声生成時に自動で保存されます。エピソード更新 API からは編集できません。
>
> **Note:** 公開状態の変更は専用エンドポイント（[エピソード公開](#エピソード公開) / [エピソード非公開](#エピソード非公開)）を使用してください。

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

## アートワーク AI 生成

```
POST /channels/:channelId/episodes/:episodeId/artwork/generate
```

エピソードのアートワークを AI（Gemini）で生成する。プロンプトを省略するとエピソードのメタデータ（タイトル・カテゴリ・説明）から自動構築する。

**リクエスト（省略可）:**
```json
{
  "prompt": "宇宙をテーマにした神秘的なデザイン。テキストは含めない。",
  "setArtwork": true
}
```

| フィールド | 型 | 必須 | デフォルト | 説明 |
|------------|-----|:----:|------------|------|
| prompt | string | | 自動生成 | 画像生成用のテキストプロンプト（1000文字以内） |
| setArtwork | bool | | true | true: 生成画像をエピソードのアートワークに自動設定 |

> **Note:** リクエストボディ全体を省略可能です。空リクエストの場合、エピソード名・カテゴリ・説明文からプロンプトを自動構築し、アートワークを自動設定します。

**レスポンス（201 Created）:**
```json
{
  "data": {
    "id": "uuid",
    "mimeType": "image/png",
    "url": "https://storage.example.com/images/xxx.png?signature=...",
    "filename": "artwork_550e8400.png",
    "fileSize": 1234567
  }
}
```

> **Note:** レスポンスは画像アップロード API（`POST /images`）と同じ形式です。`setArtwork: true` の場合、エピソードの `artwork` が自動更新されます。

**エラー（403 Forbidden）:**
```json
{
  "error": {
    "code": "FORBIDDEN",
    "message": "このエピソードのアートワーク生成権限がありません"
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

**エラー（500 Internal Server Error）:**
```json
{
  "error": {
    "code": "GENERATION_FAILED",
    "message": "画像生成に失敗しました"
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
