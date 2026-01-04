# Audio（音声生成）

## 行単位音声生成

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

---

## エピソード全体音声生成

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

# Audios（音声ファイル）

## 音声アップロード

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

---

## 音声取得

```
GET /audios/:audioId
```

---

## 音声削除

```
DELETE /audios/:audioId
```

---

# Images（画像ファイル）

## 画像アップロード

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

---

## 画像取得

```
GET /images/:imageId
```

---

## 画像削除

```
DELETE /images/:imageId
```
