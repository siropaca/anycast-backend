# Audio（音声生成）

## エピソード全体音声生成

エピソードの全台本行から音声を生成します。Gemini TTS の multi-speaker 機能を使用し、各キャラクターの声で 1 つの音声ファイルを生成します。

```
POST /channels/:channelId/episodes/:episodeId/audio/generate
```

**処理内容:**

1. エピソードの全台本行を取得（speech 行のみ対象）
2. 各キャラクターの Voice 設定を収集
3. Gemini TTS multi-speaker API で 1 つの音声ファイルを生成
4. エピソードの `fullAudio` として保存

**レスポンス:**
```json
{
  "data": {
    "id": "uuid",
    "url": "https://storage.example.com/full-episode.mp3",
    "mimeType": "audio/mpeg",
    "fileSize": 1024000,
    "durationMs": 180000
  }
}
```

**エラー:**

| コード | 説明 |
|--------|------|
| VALIDATION_ERROR | 台本に speech 行が存在しない |
| GENERATION_FAILED | 音声生成に失敗 |

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
