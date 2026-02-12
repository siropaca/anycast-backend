# Voices（ボイス）

システム管理のマスタデータ。参照のみ可能。ユーザーごとにお気に入り登録が可能。

## ボイス一覧取得

```
GET /voices
```

**クエリパラメータ:**

| パラメータ | 型 | デフォルト | 説明 |
|------------|-----|------------|------|
| provider | string | - | プロバイダでフィルタ（例: google） |
| gender | string | - | 性別でフィルタ（male / female / neutral） |

お気に入り登録済みのボイスが先頭に表示される。

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
      "sampleAudioUrl": "https://storage.example.com/...",
      "isActive": true,
      "isFavorite": true
    }
  ]
}
```

---

## ボイス取得

```
GET /voices/:voiceId
```

**レスポンス:**
```json
{
  "data": {
    "id": "uuid",
    "provider": "google",
    "providerVoiceId": "ja-JP-Wavenet-C",
    "name": "ja-JP-Wavenet-C",
    "gender": "male",
    "sampleAudioUrl": "https://storage.example.com/...",
    "isActive": true,
    "isFavorite": false
  }
}
```

---

## ボイスお気に入り登録

```
POST /voices/:voiceId/favorite
```

指定ボイスをお気に入りに登録する。既に登録済みの場合は 409 を返す。

**レスポンス（201 Created）:**
```json
{
  "data": {
    "voiceId": "uuid",
    "createdAt": "2025-01-01T00:00:00Z"
  }
}
```

**エラー（404 Not Found）:**
```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "ボイスが見つかりません"
  }
}
```

**エラー（409 Conflict）:**
```json
{
  "error": {
    "code": "ALREADY_FAVORITED",
    "message": "既にお気に入り登録済みです"
  }
}
```

---

## ボイスお気に入り解除

```
DELETE /voices/:voiceId/favorite
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

# Categories（カテゴリ）

システム管理のマスタデータ。参照のみ可能。

## カテゴリ一覧取得

```
GET /categories
```

**レスポンス:**
```json
{
  "data": [
    {
      "id": "uuid",
      "slug": "technology",
      "name": "テクノロジー",
      "image": { "id": "uuid", "url": "..." },
      "channelCount": 5,
      "episodeCount": 30,
      "sortOrder": 0,
      "isActive": true
    }
  ]
}
```

---

## カテゴリ取得（スラッグ指定）

```
GET /categories/:slug
```

**パスパラメータ:**

| パラメータ | 型 | 説明 |
|------------|-----|------|
| slug | string | カテゴリスラッグ |

**レスポンス（200 OK）:**
```json
{
  "data": {
    "id": "uuid",
    "slug": "technology",
    "name": "テクノロジー",
    "image": { "id": "uuid", "url": "..." },
    "channelCount": 5,
    "episodeCount": 30,
    "sortOrder": 0,
    "isActive": true
  }
}
```

**エラー（404 Not Found）:**
```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "カテゴリが見つかりません"
  }
}
```

