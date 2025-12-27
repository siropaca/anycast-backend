# Anycast Backend

AI 専用のポッドキャストを作成・配信できるプラットフォーム「Anycast」のバックエンド API サーバーです。

## フロントエンド

- https://github.com/siropaca/anycast-frontend

## 技術スタック

- **言語**: Go 1.24
- **フレームワーク**: Gin
- **ORM**: GORM
- **マイグレーション**: golang-migrate
- **API**: REST API
- **API ドキュメント**: Swagger (swaggo/swag)
- **DB**: PostgreSQL
- **ストレージ**: GCS（Google Cloud Storage）
- **TTS**: Google Cloud Text-to-Speech
- **バージョン管理**: mise
- **静的解析**: golangci-lint
- **ローカル環境**: Docker Compose
- **ホットリロード**: Air
- **ホスティング**: Railway

## アーキテクチャ

レイヤードアーキテクチャ + 軽量 DDD を採用しています。

```
Handler → Service → Repository → DB
```

| レイヤー | 責務 |
|----------|------|
| Handler | HTTP リクエスト/レスポンス処理 |
| Service | ビジネスロジック |
| Repository | データアクセス |
| Model | ドメインモデル |
| DTO | リクエスト/レスポンス構造体 |

詳細は [docs/adr/](docs/adr/) を参照。

## セットアップ

### 前提条件

- [mise](https://mise.jdx.dev/) がインストールされていること
- [Docker](https://www.docker.com/) および Docker Compose がインストールされていること

### インストール

```bash
# 自動セットアップ（推奨）
make bootstrap  # または make bs

# または手動でセットアップ
mise trust && mise install  # Go とツールのインストール
go mod download             # 依存関係のダウンロード
cp .env.example .env        # 環境変数ファイルの作成
```

> **Note**: シェルで `mise activate` を実行するか、`.bashrc` / `.zshrc` に設定を追加してください。

### 環境変数

| 変数 | 説明 | デフォルト |
|------|------|------------|
| `PORT` | サーバーのポート番号 | 8081 |
| `DATABASE_URL` | PostgreSQL 接続 URL | - |
| `APP_ENV` | 環境（development / production） | development |
| `JWT_SECRET` | JWT 署名用シークレットキー | - |

### DB の起動

```bash
docker compose up -d
```

### マイグレーション

```bash
make migrate-up
```

### 開発サーバーの起動

```bash
make dev
```

サーバーは http://localhost:8081 で起動します。

## API ドキュメント

### Swagger UI

開発サーバー起動後、以下の URL で Swagger UI にアクセスできます。

```
http://localhost:8081/swagger/index.html
```

API の仕様確認やインタラクティブなテストが可能です。

### ドキュメントの更新

Handler に Swagger アノテーションを追加した後、以下のコマンドでドキュメントを再生成します。

```bash
make swagger
```

### API エンドポイント

詳細は [docs/api.md](docs/api.md) を参照。

| メソッド | パス | 説明 |
|----------|------|------|
| GET | `/health` | ヘルスチェック |
| GET | `/swagger/*` | Swagger UI |
| GET | `/api/v1/voices` | ボイス一覧取得 |
| GET | `/api/v1/voices/:voiceId` | ボイス取得 |

## コマンド一覧

| コマンド | 説明 |
|----------|------|
| `make bootstrap` | 開発環境をセットアップ |
| `make dev` | 開発サーバーを起動（ホットリロード） |
| `make run` | サーバーを起動 |
| `make build` | バイナリをビルド |
| `make test` | テストを実行 |
| `make fmt` | コードをフォーマット |
| `make lint` | 静的解析を実行 |
| `make lint-fix` | 静的解析を実行（自動修正あり） |
| `make tidy` | 依存関係を整理 |
| `make clean` | ビルド成果物を削除 |
| `make swagger` | Swagger ドキュメント生成 |
| `make migrate-up` | マイグレーション実行 |
| `make migrate-down` | マイグレーションロールバック |
| `make migrate-reset` | マイグレーションリセット（down → up） |
| `make seed` | シードデータを投入（開発環境用） |
| `make token` | 開発用 JWT トークンを生成 |

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
├── railway.toml         # Railway 設定
├── nixpacks.toml        # Nixpacks ビルド設定
├── scripts/             # セットアップスクリプト
├── migrations/          # マイグレーションファイル
├── seeds/               # シードデータ（開発環境用）
├── docs/                # ドキュメント
│   ├── adr/             # Architecture Decision Records
│   ├── specification.md # 仕様書
│   ├── database.md      # DB 設計
│   └── api.md           # API 設計
├── swagger/             # Swagger ドキュメント（自動生成）
├── http/                # HTTP リクエストファイル
├── internal/            # 内部パッケージ
│   ├── apperror/        # カスタムエラー型
│   ├── config/          # 設定管理
│   ├── db/              # DB 接続
│   ├── di/              # DI コンテナ
│   ├── dto/             # Data Transfer Objects
│   ├── handler/         # ハンドラー
│   ├── logger/          # 構造化ログ
│   ├── middleware/      # ミドルウェア
│   ├── model/           # ドメインモデル
│   ├── repository/      # データアクセス層
│   ├── router/          # ルーティング
│   └── service/         # ビジネスロジック層
├── README.md
└── CLAUDE.md
```
