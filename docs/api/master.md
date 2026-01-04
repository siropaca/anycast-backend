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
      "sortOrder": 0,
      "isActive": true
    }
  ]
}
```

---

# Sound Effects（効果音）

システム管理のマスタデータ。参照のみ可能。

## 効果音一覧取得

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

---

## 効果音取得

```
GET /sound-effects/:sfxId
```
