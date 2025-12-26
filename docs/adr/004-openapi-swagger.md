# ADR-004: OpenAPI (Swagger) による API ドキュメント

## ステータス

Accepted

## コンテキスト

フロントエンド開発者や外部連携のために、API ドキュメントを整備する必要があった。

以下の要件を考慮した:
- API 仕様の可視化
- インタラクティブなテスト機能
- コードとドキュメントの同期
- 開発効率の向上

## 決定

**swaggo/swag** を使用して、コードから OpenAPI 仕様を自動生成する。

```go
// @Summary ボイス一覧取得
// @Description 利用可能なボイスの一覧を取得します
// @Tags voices
// @Param provider query string false "プロバイダでフィルタ"
// @Success 200 {object} map[string][]response.VoiceResponse
// @Router /voices [get]
func (h *VoiceHandler) ListVoices(c *gin.Context) {
```

## 選択肢

### 選択肢 1: swaggo/swag（コードファースト）

- Go コードのコメントから仕様を生成
- コードと仕様が同期しやすい
- Gin との統合が容易
- アノテーションの学習コスト

### 選択肢 2: OpenAPI スキーマファースト

- 仕様を先に定義
- フロントエンドとの並行開発が可能
- コードとの同期が手動
- 二重管理のリスク

### 選択肢 3: 手動ドキュメント（Markdown）

- シンプル
- 自由度が高い
- コードとの同期が困難
- インタラクティブなテスト不可

### 選択肢 4: Postman / Insomnia

- インタラクティブなテストが可能
- チーム共有が容易
- API 仕様としては不完全
- 外部サービス依存

## 理由

1. **DRY 原則**: コードが唯一の信頼源となる
2. **自動同期**: コード変更時に `make swagger` で仕様を更新
3. **開発者体験**: Swagger UI でインタラクティブにテスト可能
4. **エコシステム**: Go + Gin での実績が豊富

## 結果

- Handler に Swagger アノテーションを追加
- `make swagger` で `docs/` ディレクトリにドキュメント生成
- `/swagger/*` で Swagger UI にアクセス可能
- API 変更時はアノテーションの更新と `make swagger` の実行が必要
