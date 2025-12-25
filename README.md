# Anycast Backend

AI 専用のポッドキャストを作成・配信できるプラットフォーム「Anycast」のバックエンド API サーバーです。

## フロントエンド

- https://github.com/siropaca/anycast-forntend

## 技術スタック

- **言語**: Go 1.24
- **フレームワーク**: Gin
- **API**: REST API
- **DB**: PostgreSQL
- **ストレージ**: GCS（Google Cloud Storage）
- **TTS**: Google Cloud Text-to-Speech
- **バージョン管理**: mise
- **ローカル環境**: Docker Compose
- **ホットリロード**: Air
- **ホスティング**: Railway

## セットアップ

### 前提条件

- [mise](https://mise.jdx.dev/) がインストールされていること
- [Docker](https://www.docker.com/) および Docker Compose がインストールされていること

### インストール

```bash
# mise でツールのバージョンを設定
mise trust && mise install

# 依存関係のインストール
go mod download

# 環境変数の設定
cp .env.example .env
```

### DB の起動

```bash
docker compose up -d
```

### 開発サーバーの起動

```bash
make dev
```

サーバーは http://localhost:8081 で起動します。

## API エンドポイント

詳細は [doc/api.md](doc/api.md) を参照。

| メソッド | パス | 説明 |
|----------|------|------|
| GET | `/health` | ヘルスチェック |

## コマンド一覧

| コマンド | 説明 |
|----------|------|
| `make dev` | 開発サーバーを起動（ホットリロード） |
| `make run` | サーバーを起動 |
| `make build` | バイナリをビルド |
| `make test` | テストを実行 |
| `make fmt` | コードをフォーマット |
| `make lint` | 静的解析を実行 |
| `make tidy` | 依存関係を整理 |
| `make clean` | ビルド成果物を削除 |

## ディレクトリ構成

```
.
├── main.go              # エントリーポイント
├── go.mod
├── go.sum
├── Makefile             # コマンド定義
├── .env.example         # 環境変数のサンプル
├── .air.toml            # Air 設定
├── .mise.toml           # mise 設定
├── docker-compose.yml   # ローカル DB
├── doc/                 # ドキュメント
│   ├── specification.md # 仕様書
│   ├── database.md      # DB 設計
│   └── api.md           # API 設計
├── README.md
└── CLAUDE.md
```
