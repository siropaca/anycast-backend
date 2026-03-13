# テスト・手動検証ガイド

## ユニットテスト

- 外部依存のないユニットテストは、実装時に必ず作成する
- テスト実行: `make test`

---

## API の手動テスト

### 前提条件

- Docker で DB が起動していること（`docker compose up -d`）
- シードデータが投入済みであること（`make seed`）

### 手順

1. **サーバー起動**: `make dev`（ホットリロード）または `make run`
2. **JWT トークン取得**: `make token` を実行し、出力されたトークンを使用する
3. **API 呼び出し**: `curl` で実行する

### シードデータのテスト用 ID

| リソース | ID | 説明 |
|----------|------|------|
| test_user | `8def69af-dae9-4641-a0e5-100107626933` | `make token` で生成されるユーザー |
| channel (テックトーク) | `ea9a266e-f532-417c-8916-709d0233941c` | test_user 所有、キャラクター2名 |
| episode (AI の未来を語る) | `eb960304-f86e-4364-be5d-d3d5126c9601` | テックトーク内のエピソード |

### curl の例

```bash
# トークン取得
TOKEN=$(make token 2>&1)

# 台本生成（非同期）
curl -s -X POST "http://localhost:8081/api/v1/channels/ea9a266e-f532-417c-8916-709d0233941c/episodes/eb960304-f86e-4364-be5d-d3d5126c9601/script/generate-async" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"prompt":"AIの未来について","durationMinutes":3,"withEmotion":true}' | jq .

# ジョブ状態確認
curl -s "http://localhost:8081/api/v1/script-jobs/{JOB_ID}" \
  -H "Authorization: Bearer $TOKEN" | jq .

# 自分のジョブ一覧
curl -s "http://localhost:8081/api/v1/me/script-jobs" \
  -H "Authorization: Bearer $TOKEN" | jq .

# 自分のジョブ一覧（ステータスフィルタ）
curl -s "http://localhost:8081/api/v1/me/script-jobs?status=completed" \
  -H "Authorization: Bearer $TOKEN" | jq .
```

### 開発用エンドポイント（認証不要）

DB やシードデータ不要で台本生成を直接テストできる。`APP_ENV=development` のときのみ有効。

```bash
curl -s -X POST "http://localhost:8081/dev/script/generate" \
  -H "Content-Type: application/json" \
  -d '{
    "episodeTitle": "テスト回",
    "durationMinutes": 3,
    "episodeNumber": 1,
    "channelName": "テストチャンネル",
    "channelCategory": "テクノロジー",
    "characters": [
      {"name": "太郎", "gender": "male", "persona": "テック好きの大学生"},
      {"name": "花子", "gender": "female", "persona": "AI研究者"}
    ],
    "theme": "最近のAI技術について",
    "withEmotion": true
  }' | jq .
```

- 同期処理のためレスポンスまで数十秒かかる

### デバッグ出力（`tmp/`）

ローカル開発時、以下のディレクトリにデバッグ用ファイルが出力される（`tmp/` は `.gitignore` に含まれる）。

| ディレクトリ | 内容 | 出力条件 |
|-------------|------|----------|
| `tmp/traces/{episodeTitle}/` | 台本生成の Phase ごとのトレース（Markdown）。ブリーフ JSON、プロンプト、LLM レスポンスなどを含む | `TRACE_MODE=file` |
| `tmp/audio-debug/{jobID}/` | TTS 完了直後のスピーカー別オリジナル音源（WAV） | `APP_ENV != production` |

### 注意事項

- `make token` の出力には stderr の情報が混ざる場合があるため、`TOKEN=$(make token 2>&1)` で取得する
- 台本生成は非同期処理のため、レスポンスで返る `jobId` を使ってジョブ状態をポーリングする
- ローカル環境では Cloud Tasks がないため goroutine で直接実行される
- **MCP ツール（`mcp__anycast__*`）は本番環境に接続されているため、動作確認やテスト目的で使用しない。** 開発環境でのテストは上記の curl による直接 API 呼び出しで行うこと
