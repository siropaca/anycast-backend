# Anycast Backend

AI を活用してポッドキャストを作成・配信できるプラットフォーム「Anycast」のバックエンド API サーバーです。

## フロントエンド

- https://github.com/siropaca/anycast-forntend

## 技術スタック

- **言語**: Go
- **API**: REST API
- **DB**: PostgreSQL
- **バージョン管理**: mise
- **ローカル環境**: Docker Compose
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
```

### DB の起動

```bash
docker compose up -d
```

### 開発サーバーの起動

```bash
go run main.go
```

## ディレクトリ構成

```
.
├── README.md
├── CLAUDE.md
├── .mise.toml
├── docker-compose.yml
├── go.mod
└── main.go
```
