# Anycast API 設計

## 概要

- **ベース URL**: `/api/v1`
- **形式**: REST API
- **データ形式**: JSON
- **認証**: Bearer Token（予定）

---

## 共通仕様

### レスポンス形式

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

### ページネーション

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

### 権限

| 権限レベル | 説明 |
|------------|------|
| Public | 誰でも閲覧可能 |
| Owner | 自身のリソースのみ操作可能 |
| Admin | 運営のみ操作可能 |

**エンドポイント別権限:**

| リソース | 参照 | 作成 | 更新 | 削除 |
|----------|:----:|:----:|:----:|:----:|
| Channels | Public | Owner | Owner | Owner |
| Characters | Public | Owner | Owner | Owner |
| Episodes | Public | Owner | Owner | Owner |
| Script / ScriptLines | Public | Owner | Owner | Owner |
| Audio（生成） | - | Owner | - | - |
| Audios（アップロード） | Owner | Owner | - | Owner |
| Images（アップロード） | Owner | Owner | - | Owner |
| Voices | Public | Admin | Admin | Admin |
| Sound Effects | Public | Admin | Admin | Admin |

---

## Channels

### チャンネル一覧取得

```
GET /channels
```

**クエリパラメータ:** ページネーション

**レスポンス:**
```json
{
  "data": [
    {
      "id": "uuid",
      "name": "チャンネル名",
      "description": "説明",
      "artwork": { "id": "uuid", "url": "..." },
      "createdAt": "2025-01-01T00:00:00Z",
      "updatedAt": "2025-01-01T00:00:00Z"
    }
  ]
}
```

### チャンネル取得

```
GET /channels/:channelId
```

**レスポンス:**
```json
{
  "data": {
    "id": "uuid",
    "name": "チャンネル名",
    "description": "説明",
    "artwork": { "id": "uuid", "url": "..." },
    "characters": [
      {
        "id": "uuid",
        "name": "太郎",
        "persona": "明るい性格",
        "voice": {
          "id": "uuid",
          "name": "ja-JP-Wavenet-C",
          "provider": "google",
          "gender": "male"
        }
      }
    ],
    "createdAt": "2025-01-01T00:00:00Z",
    "updatedAt": "2025-01-01T00:00:00Z"
  }
}
```

### チャンネル作成

```
POST /channels
```

**リクエスト:**
```json
{
  "name": "チャンネル名",
  "description": "説明",
  "artworkImageId": "uuid"
}
```

### チャンネル更新

```
PATCH /channels/:channelId
```

**リクエスト:**
```json
{
  "name": "新しいチャンネル名",
  "description": "新しい説明",
  "artworkImageId": "uuid"
}
```

### チャンネル削除

```
DELETE /channels/:channelId
```

---

## Characters

### キャラクター一覧取得

```
GET /channels/:channelId/characters
```

### キャラクター作成

```
POST /channels/:channelId/characters
```

**リクエスト:**
```json
{
  "name": "太郎",
  "persona": "明るく元気な性格。語尾に「だよね」をつける。",
  "voiceId": "uuid"
}
```

**バリデーション:**
- name: 必須、同一チャンネル内で一意、`__` で始まる名前は禁止
- voiceId: 必須、is_active = true のボイスのみ指定可能

### キャラクター更新

```
PATCH /channels/:channelId/characters/:characterId
```

**リクエスト:**
```json
{
  "name": "新しい名前",
  "persona": "新しいペルソナ",
  "voiceId": "uuid"
}
```

### キャラクター削除

```
DELETE /channels/:channelId/characters/:characterId
```

---

## Episodes

### エピソード一覧取得

```
GET /channels/:channelId/episodes
```

**クエリパラメータ:** ページネーション

### エピソード取得

```
GET /channels/:channelId/episodes/:episodeId
```

**レスポンス:**
```json
{
  "data": {
    "id": "uuid",
    "title": "エピソードタイトル",
    "description": "エピソードの説明",
    "bgm": { "id": "uuid", "url": "..." },
    "fullAudio": { "id": "uuid", "url": "..." },
    "script": [
      {
        "id": "uuid",
        "lineOrder": 0,
        "lineType": "speech",
        "speaker": { "id": "uuid", "name": "太郎" },
        "text": "こんにちは",
        "emotion": null,
        "audio": { "id": "uuid", "url": "..." }
      },
      {
        "id": "uuid",
        "lineOrder": 1,
        "lineType": "silence",
        "durationMs": 800
      },
      {
        "id": "uuid",
        "lineOrder": 2,
        "lineType": "sfx",
        "sfx": { "id": "uuid", "name": "chime" },
        "volume": 0.8
      }
    ],
    "createdAt": "2025-01-01T00:00:00Z",
    "updatedAt": "2025-01-01T00:00:00Z"
  }
}
```

### エピソード作成

```
POST /channels/:channelId/episodes
```

**リクエスト:**
```json
{
  "title": "エピソードタイトル",
  "description": "エピソードの説明",
  "bgmAudioId": "uuid"
}
```

### エピソード更新

```
PATCH /channels/:channelId/episodes/:episodeId
```

**リクエスト:**
```json
{
  "title": "新しいタイトル",
  "description": "新しい説明",
  "bgmAudioId": "uuid"
}
```

### エピソード削除

```
DELETE /channels/:channelId/episodes/:episodeId
```

---

## Script（台本）

### 台本テキスト取り込み

```
POST /channels/:channelId/episodes/:episodeId/script/import
```

**リクエスト:**
```json
{
  "text": "太郎: こんにちは\n花子: やあ\n__SILENCE__: 800\n__SFX__: chime"
}
```

**レスポンス（成功時）:**
```json
{
  "data": {
    "lines": [
      { "id": "uuid", "lineOrder": 0, "lineType": "speech", ... }
    ]
  }
}
```

**レスポンス（エラー時）:**
```json
{
  "error": {
    "code": "SCRIPT_PARSE_ERROR",
    "message": "台本のパースに失敗しました",
    "details": [
      { "line": 3, "reason": "不明な話者: 三郎" },
      { "line": 5, "reason": "__SFX__ の値が不正です" }
    ]
  }
}
```

### 台本テキスト出力

```
GET /channels/:channelId/episodes/:episodeId/script/export
```

**レスポンス:**
```json
{
  "data": {
    "text": "太郎: こんにちは\n花子: やあ\n__SILENCE__: 800"
  }
}
```

### 台本を AI で生成

```
POST /channels/:channelId/episodes/:episodeId/script/generate
```

**リクエスト:**
```json
{
  "prompt": "今日の天気について楽しく話す"
}
```

- `prompt`: テーマやシナリオを入力。URL が含まれていれば RAG で内容を取得して台本生成に利用

**レスポンス:**
```json
{
  "data": {
    "lines": [ ... ]
  }
}
```

---

## ScriptLines（台本行）

### 行追加

```
POST /channels/:channelId/episodes/:episodeId/script/lines
```

**リクエスト（speech）:**
```json
{
  "lineType": "speech",
  "speakerId": "uuid",
  "text": "こんにちは",
  "emotion": "嬉しい",
  "insertAfter": "uuid"
}
```

- `speakerId`: 同じ Channel に属する Character の ID を指定
- `emotion`: 感情・喋り方の指定（任意）

**リクエスト（silence）:**
```json
{
  "lineType": "silence",
  "durationMs": 800,
  "insertAfter": "uuid"
}
```

**リクエスト（sfx）:**
```json
{
  "lineType": "sfx",
  "sfxId": "uuid",
  "volume": 0.8,
  "insertAfter": "uuid"
}
```

- `insertAfter`: 指定した行の後に挿入。null の場合は先頭に挿入。

### 行更新

```
PATCH /channels/:channelId/episodes/:episodeId/script/lines/:lineId
```

**リクエスト:**
```json
{
  "text": "新しいセリフ",
  "emotion": "笑いながら"
}
```

### 行削除

```
DELETE /channels/:channelId/episodes/:episodeId/script/lines/:lineId
```

### 行並び替え

```
POST /channels/:channelId/episodes/:episodeId/script/reorder
```

**リクエスト:**
```json
{
  "lineIds": ["uuid1", "uuid2", "uuid3"]
}
```

---

## Audio（音声生成）

### 行単位音声生成

```
POST /channels/:channelId/episodes/:episodeId/script/lines/:lineId/audio/generate
```

**レスポンス:**
```json
{
  "data": {
    "audio": {
      "id": "uuid",
      "url": "https://storage.example.com/audio.mp3",
      "durationMs": 2500
    }
  }
}
```

### エピソード全体音声生成

```
POST /channels/:channelId/episodes/:episodeId/audio/generate
```

**レスポンス:**
```json
{
  "data": {
    "audio": {
      "id": "uuid",
      "url": "https://storage.example.com/full-episode.mp3",
      "durationMs": 180000
    }
  }
}
```

---

## Audios（音声ファイル）

### 音声アップロード

```
POST /audios
```

**リクエスト:** `multipart/form-data`

| フィールド | 型 | 必須 | 説明 |
|------------|-----|:----:|------|
| file | File | ◯ | アップロードするファイル（mp3 など） |

**レスポンス:**
```json
{
  "data": {
    "id": "uuid",
    "mimeType": "audio/mpeg",
    "url": "https://storage.example.com/file.mp3",
    "filename": "bgm.mp3",
    "fileSize": 1024000,
    "durationMs": 180000
  }
}
```

### 音声取得

```
GET /audios/:audioId
```

### 音声削除

```
DELETE /audios/:audioId
```

---

## Images（画像ファイル）

### 画像アップロード

```
POST /images
```

**リクエスト:** `multipart/form-data`

| フィールド | 型 | 必須 | 説明 |
|------------|-----|:----:|------|
| file | File | ◯ | アップロードするファイル（png, jpeg など） |

**レスポンス:**
```json
{
  "data": {
    "id": "uuid",
    "mimeType": "image/png",
    "url": "https://storage.example.com/artwork.png",
    "filename": "artwork.png",
    "fileSize": 512000
  }
}
```

### 画像取得

```
GET /images/:imageId
```

### 画像削除

```
DELETE /images/:imageId
```

---

## Voices（ボイス）

システム管理のマスタデータ。参照のみ可能。

### ボイス一覧取得

```
GET /voices
```

**クエリパラメータ:**

| パラメータ | 型 | デフォルト | 説明 |
|------------|-----|------------|------|
| provider | string | - | プロバイダでフィルタ（例: google） |
| gender | string | - | 性別でフィルタ（male / female / neutral） |
| activeOnly | bool | true | true の場合 is_active = true のみ取得 |

**レスポンス:**
```json
{
  "data": [
    {
      "id": "uuid",
      "provider": "google",
      "providerVoiceId": "ja-JP-Wavenet-C",
      "name": "ja-JP-Wavenet-C",
      "gender": "male",
      "isActive": true
    }
  ]
}
```

### ボイス取得

```
GET /voices/:voiceId
```

---

## Sound Effects（効果音）

システム管理のマスタデータ。参照のみ可能。

### 効果音一覧取得

```
GET /sound-effects
```

**レスポンス:**
```json
{
  "data": [
    {
      "id": "uuid",
      "name": "chime",
      "description": "チャイム音",
      "audio": { "id": "uuid", "url": "...", "durationMs": 1500 }
    }
  ]
}
```

### 効果音取得

```
GET /sound-effects/:sfxId
```

---

## エラーコード一覧

| コード | HTTP Status | 説明 |
|--------|-------------|------|
| NOT_FOUND | 404 | リソースが見つからない |
| VALIDATION_ERROR | 400 | バリデーションエラー |
| DUPLICATE_NAME | 409 | 名前が重複している |
| RESERVED_NAME | 400 | 予約語を使用している |
| SCRIPT_PARSE_ERROR | 400 | 台本のパースに失敗 |
| GENERATION_FAILED | 500 | 音声/台本の生成に失敗 |
| MEDIA_UPLOAD_FAILED | 500 | メディアアップロードに失敗 |
| SFX_IN_USE | 409 | 効果音が使用中のため削除不可 |
