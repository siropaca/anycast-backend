# Voices（ボイス）

システム管理のマスタデータ。参照のみ可能。

## ボイス一覧取得

```
GET /voices
```

**クエリパラメータ:**

| パラメータ | 型 | デフォルト | 説明 |
|------------|-----|------------|------|
| provider | string | - | プロバイダでフィルタ（例: google） |
| gender | string | - | 性別でフィルタ（male / female / neutral） |

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
      "isActive": true
    }
  ]
}
```

---

## ボイス取得

```
GET /voices/:voiceId
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

