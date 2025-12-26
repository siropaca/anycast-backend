#!/usr/bin/env bash
# 開発環境セットアップスクリプト

set -euo pipefail

# 色付きメッセージ用の関数
print_info() { echo -e "\033[34m[INFO]\033[0m $1"; }
print_success() { echo -e "\033[32m[SUCCESS]\033[0m $1"; }
print_error() { echo -e "\033[31m[ERROR]\033[0m $1"; }

print_info "Anycast Backend セットアップを開始します"

# 1. mise のセットアップ
print_info "mise でツールをインストールしています..."
if ! command -v mise &> /dev/null; then
    print_error "mise がインストールされていません。https://mise.jdx.dev/ からインストールしてください。"
    exit 1
fi
mise trust && mise install
print_success "mise のツールがインストールされました"

# 2. Go 依存関係のインストール
print_info "Go の依存関係をインストール中..."
go mod download
print_success "Go の依存関係がインストールされました"

# 3. 環境変数ファイルのセットアップ
if [ ! -f .env ]; then
    print_info ".env ファイルを作成中..."
    cp .env.example .env
    print_success ".env ファイルが作成されました"
else
    print_info ".env ファイルは既に存在します"
fi

# 4. Docker Compose の確認
if command -v docker compose &> /dev/null; then
    print_info "Docker Compose でデータベースを起動しますか？ (y/N)"
    read -r response
    if [[ "$response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
        docker compose up -d
        print_success "データベースが起動しました"

        # DB の起動を少し待つ
        print_info "データベースの準備を待っています..."
        sleep 3

        # マイグレーションの実行
        print_info "マイグレーションを実行しますか？ (y/N)"
        read -r response
        if [[ "$response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
            mise exec -- make migrate-up
            print_success "マイグレーションが完了しました"
        fi
    fi
else
    print_info "Docker Compose がインストールされていません。手動でデータベースをセットアップしてください。"
fi

print_success "セットアップが完了しました！"
echo ""
echo "インストールされたツール: go, air, swag, golangci-lint, migrate"
echo ""
print_info "開発サーバーを起動するには 'make dev' を実行してください"
