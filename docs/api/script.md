# Script（台本）

## 台本を AI で生成

```
POST /channels/:channelId/episodes/:episodeId/script/generate
```

**リクエスト:**
```json
{
  "prompt": "今日の天気について楽しく話す",
  "durationMinutes": 10,
  "withEmotion": true
}
```

| フィールド | 型 | 必須 | 説明 |
|------------|-----|:----:|------|
| prompt | string | ◯ | テーマやシナリオ（2000文字以内）。URL が含まれていれば RAG で内容を取得して台本生成に利用 |
| durationMinutes | int | | エピソードの長さ（分）。3〜30の範囲で指定。デフォルト: 10 |
| withEmotion | bool | | 感情を付与するかどうか。デフォルト: false |

> **Note:** `prompt` はエピソードの `userPrompt` として自動保存されます。

**レスポンス:**
```json
{
  "data": {
    "lines": [ ... ]
  }
}
```

---

## 台本テキスト取り込み

```
POST /channels/:channelId/episodes/:episodeId/script/import
```

テキスト形式の台本をインポートする。既存の台本がある場合は全て削除される。

**リクエスト:**
```json
{
  "text": "太郎: こんにちは\n花子: [嬉しそうに] やあ\n__SILENCE__: 800\n__SFX__: chime"
}
```

| フィールド | 型 | 必須 | 説明 |
|------------|-----|:----:|------|
| text | string | ◯ | 台本テキスト |

**テキストフォーマット:**

| 行タイプ | 形式 | 例 |
|----------|------|-----|
| speech | `話者名: [感情] セリフ` | `太郎: [嬉しそうに] こんにちは` |
| silence | `__SILENCE__: ミリ秒` | `__SILENCE__: 800` |
| sfx | `__SFX__: 効果音名` | `__SFX__: chime` |

- `[感情]` は省略可能
- 話者名はチャンネルに登録されているキャラクター名のみ使用可能
- 効果音名は登録済みの効果音名のみ使用可能

**レスポンス（成功時）:**
```json
{
  "data": [
    { "id": "uuid", "lineOrder": 0, "lineType": "speech", ... }
  ]
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

---

## 台本テキスト出力

```
GET /channels/:channelId/episodes/:episodeId/script/export
```

台本をテキストファイルとしてダウンロードする。出力されたテキストはそのままインポート可能。

**レスポンス:**
- Content-Type: `text/plain; charset=utf-8`
- Content-Disposition: `attachment; filename="エピソード名.txt"; filename*=UTF-8''...`

```
太郎: こんにちは
花子: [嬉しそうに] やあ
__SILENCE__: 800
__SFX__: chime
```

---

# ScriptLines（台本行）

## 台本行一覧取得

```
GET /channels/:channelId/episodes/:episodeId/script/lines
```

指定したエピソードの台本行一覧を `lineOrder` 順で取得する。

**レスポンス:**
```json
{
  "data": [
    {
      "id": "uuid",
      "lineOrder": 0,
      "lineType": "speech",
      "speaker": { "id": "uuid", "name": "太郎" },
      "text": "こんにちは",
      "emotion": null,
      "audio": { "id": "uuid", "url": "...", "durationMs": 2500 },
      "createdAt": "2025-01-01T00:00:00Z",
      "updatedAt": "2025-01-01T00:00:00Z"
    },
    {
      "id": "uuid",
      "lineOrder": 1,
      "lineType": "silence",
      "durationMs": 800,
      "createdAt": "2025-01-01T00:00:00Z",
      "updatedAt": "2025-01-01T00:00:00Z"
    },
    {
      "id": "uuid",
      "lineOrder": 2,
      "lineType": "sfx",
      "sfx": { "id": "uuid", "name": "chime" },
      "volume": 0.8,
      "createdAt": "2025-01-01T00:00:00Z",
      "updatedAt": "2025-01-01T00:00:00Z"
    }
  ]
}
```

---

## 行追加

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

---

## 行更新

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

---

## 行削除

```
DELETE /channels/:channelId/episodes/:episodeId/script/lines/:lineId
```

指定した台本行を削除する。

**レスポンス:**
- `204 No Content`: 削除成功
- `403 Forbidden`: チャンネルのオーナーでない場合
- `404 Not Found`: 台本行が存在しない場合

---

## 行並び替え

```
POST /channels/:channelId/episodes/:episodeId/script/reorder
```

**リクエスト:**
```json
{
  "lineIds": ["uuid1", "uuid2", "uuid3"]
}
```
