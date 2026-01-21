# Audio（音声生成）

## エピソード全体音声生成

エピソードの全台本行から音声を生成します。Gemini TTS の multi-speaker 機能を使用し、各キャラクターの声で 1 つの音声ファイルを生成します。

```
POST /channels/:channelId/episodes/:episodeId/audio/generate
```

**リクエスト:**
```json
{
  "voiceStyle": "Read aloud in a warm, welcoming tone"
}
```

| フィールド | 型 | 必須 | 説明 |
|------------|-----|:----:|------|
| voiceStyle | string | | 音声生成のスタイル指示（500文字以内） |

**処理内容:**

1. エピソードの全台本行を取得
2. 各キャラクターの Voice 設定を収集
3. リクエストの `voiceStyle` をスタイル指示として適用
4. Gemini TTS multi-speaker API で 1 つの音声ファイルを生成
5. エピソードの `fullAudio` として保存
6. `voiceStyle` をエピソードに保存

**voiceStyle について:**

リクエストで `voiceStyle` が指定された場合、音声生成時のスタイル指示として使用され、エピソードに保存されます。

例: `"Read aloud in a warm, welcoming tone"`

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

## 音声アップロード

```
POST /audios
```

**リクエスト:** `multipart/form-data`

| フィールド | 型 | 必須 | 説明 |
|------------|-----|:----:|------|
| file | File | ◯ | アップロードする音声ファイル（mp3, wav, ogg, aac, m4a） |

**レスポンス:**
```json
{
  "data": {
    "id": "uuid",
    "mimeType": "audio/mpeg",
    "url": "https://storage.example.com/audio.mp3",
    "filename": "bgm.mp3",
    "fileSize": 1024000,
    "durationMs": 180000
  }
}
```

> **Note:** `durationMs` は MP3 形式の場合のみビットレートベースで推定されます。その他の形式では 0 が返されます。

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
