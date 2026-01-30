# Recommendations（おすすめ）

パーソナライズされたおすすめコンテンツを取得する API。ログイン状態によって挙動が変わる。

## 概要

| 状態 | 挙動 |
|------|------|
| 未ログイン | 人気順・新着順に基づくおすすめ |
| ログイン | ユーザーの再生履歴・プレイリストに基づくパーソナライズ |

---

## おすすめチャンネル取得

```
GET /recommendations/channels
```

**権限:** Optional（ログイン任意）

### クエリパラメータ

| パラメータ | 型 | デフォルト | 説明 |
|------------|-----|------------|------|
| limit | int | 20 | 取得件数（1〜50） |
| offset | int | 0 | オフセット |
| categoryId | uuid | - | カテゴリ ID でフィルタ |

### おすすめロジック

#### 未ログイン時

1. **対象**: 公開中チャンネル
2. **スコア計算**:
   - 総再生回数（チャンネル内全エピソードの合計）
   - 新着度（最新エピソードの公開日）
3. **多様性**: 同一カテゴリは最大 3 件まで連続しない

#### ログイン時

1. **対象**: 公開中チャンネル（自分のチャンネルは除外）
2. **パーソナライズスコア**:
   - 再生履歴のカテゴリ傾向に一致 → 加点
   - 「後で聴く」に追加したエピソードのチャンネルカテゴリ → 加点
   - 既に再生したエピソードがあるチャンネル → 減点（新規チャンネル発見を促進）
3. **多様性**: 同一カテゴリは最大 3 件まで連続しない

### レスポンス

```json
{
  "data": [
    {
      "id": "uuid",
      "name": "チャンネル名",
      "description": "チャンネルの説明",
      "artwork": { "id": "uuid", "url": "https://..." },
      "category": { "id": "uuid", "slug": "technology", "name": "テクノロジー" },
      "episodeCount": 10,
      "totalPlayCount": 1500,
      "latestEpisodeAt": "2025-01-20T10:00:00Z"
    }
  ],
  "pagination": {
    "total": 100,
    "limit": 20,
    "offset": 0
  }
}
```

| フィールド | 型 | 説明 |
|------------|-----|------|
| id | uuid | チャンネル ID |
| name | string | チャンネル名 |
| description | string | チャンネルの説明 |
| artwork | object \| null | アートワーク情報 |
| category | object | カテゴリ情報 |
| episodeCount | int | 公開中エピソード数 |
| totalPlayCount | int | チャンネル内全エピソードの総再生回数 |
| latestEpisodeAt | datetime \| null | 最新エピソードの公開日時 |

---

## おすすめエピソード取得

```
GET /recommendations/episodes
```

**権限:** Optional（ログイン任意）

### クエリパラメータ

| パラメータ | 型 | デフォルト | 説明 |
|------------|-----|------------|------|
| limit | int | 20 | 取得件数（1〜50） |
| offset | int | 0 | オフセット |
| categoryId | uuid | - | カテゴリ ID でフィルタ（チャンネルのカテゴリ） |

### おすすめロジック

#### 未ログイン時

1. **対象**: 公開中エピソード
2. **スコア計算**:
   - 再生回数
   - 新着度（公開日）
3. **多様性**: 同一チャンネルは最大 2 件まで連続しない

#### ログイン時

1. **対象**: 公開中エピソード（自分のエピソードは除外）
2. **優先度の高い順に混合**:
   - **途中再生中**: 再生履歴があり `completed=false` のエピソード（最大 3 件）
   - **「後で聴く」**: デフォルトプレイリストに追加済みで未再生のエピソード（最大 3 件）
   - **パーソナライズ**: 再生履歴のカテゴリ傾向に基づくエピソード
3. **除外**: 既に `completed=true` のエピソード
4. **多様性**: 同一チャンネルは最大 2 件まで連続しない

### レスポンス

```json
{
  "data": [
    {
      "id": "uuid",
      "title": "エピソードタイトル",
      "description": "エピソードの説明",
      "artwork": { "id": "uuid", "url": "https://..." },
      "fullAudio": { "id": "uuid", "url": "https://...", "durationMs": 180000 },
      "playCount": 500,
      "publishedAt": "2025-01-20T10:00:00Z",
      "channel": {
        "id": "uuid",
        "name": "チャンネル名",
        "artwork": { "id": "uuid", "url": "https://..." },
        "category": { "id": "uuid", "slug": "technology", "name": "テクノロジー" }
      },
      "playbackProgress": {
        "progressMs": 60000,
        "completed": false
      },
      "inListenLater": true
    }
  ],
  "pagination": {
    "total": 100,
    "limit": 20,
    "offset": 0
  }
}
```

| フィールド | 型 | 説明 |
|------------|-----|------|
| id | uuid | エピソード ID |
| title | string | エピソードタイトル |
| description | string | エピソードの説明 |
| artwork | object \| null | アートワーク情報（エピソード固有、なければ null） |
| fullAudio | object \| null | 音声ファイル情報 |
| playCount | int | 再生回数 |
| publishedAt | datetime | 公開日時 |
| channel | object | チャンネル情報（カテゴリ含む） |
| playbackProgress | object \| null | 再生進捗（ログイン時のみ、履歴がある場合） |
| inListenLater | bool | 「後で聴く」に追加済みか（ログイン時のみ） |

### playbackProgress

| フィールド | 型 | 説明 |
|------------|-----|------|
| progressMs | int | 再生位置（ミリ秒） |
| completed | bool | 再生完了フラグ |

---

## 実装メモ

### 任意認証（Optional Auth）

これらのエンドポイントは「任意認証」を使用する。

- トークンがある場合: 検証してユーザー ID をコンテキストにセット
- トークンがない場合: 未ログインとして処理を継続（エラーにしない）
- トークンが不正な場合: 401 Unauthorized を返す

### 多様性の実装

結果リストを構築する際、連続する同一カテゴリ/チャンネルの件数をカウントし、上限に達した場合は次の候補をスキップして別のカテゴリ/チャンネルのアイテムを挿入する。
