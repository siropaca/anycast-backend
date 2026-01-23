#!/usr/bin/env bash
# 開発環境セットアップスクリプト

set -euo pipefail

# 色付きメッセージ用の関数
print_info() { echo -e "\033[34m[INFO]\033[0m $1"; }
print_success() { echo -e "\033[32m[SUCCESS]\033[0m $1"; }
print_error() { echo -e "\033[31m[ERROR]\033[0m $1"; }

print_info "Anycast Backend セットアップを開始します"

# 1. FFmpeg のセットアップ（Homebrew）
if ! command -v ffmpeg &> /dev/null; then
    print_info "FFmpeg がインストールされていません"
    if command -v brew &> /dev/null; then
        print_info "Homebrew で FFmpeg をインストールしますか？ (y/N)"
        read -r response
        if [[ "$response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
            brew install ffmpeg
            print_success "FFmpeg がインストールされました"
        else
            print_info "FFmpeg は手動でインストールしてください（音声ミキシング機能で使用）"
        fi
    else
        print_info "Homebrew がインストールされていません。FFmpeg を手動でインストールしてください。"
        print_info "  macOS: brew install ffmpeg"
        print_info "  Ubuntu: sudo apt install ffmpeg"
    fi
else
    print_success "FFmpeg は既にインストールされています: $(ffmpeg -version 2>&1 | head -1)"
fi

# 2. mise のセットアップ
print_info "mise でツールをインストールしています..."
if ! command -v mise &> /dev/null; then
    print_error "mise がインストールされていません。https://mise.jdx.dev/ からインストールしてください。"
    exit 1
fi
mise trust && mise install
print_success "mise のツールがインストールされました"

# 3. Go 依存関係のインストール
print_info "Go の依存関係をインストール中..."
go mod download
print_success "Go の依存関係がインストールされました"

# 4. 環境変数ファイルのセットアップ
if [ ! -f .env ]; then
    print_info ".env ファイルを作成中..."
    cp .env.example .env
    print_success ".env ファイルが作成されました"
else
    print_info ".env ファイルは既に存在します"
fi

# 5. Docker Compose の確認
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
echo "インストールされたツール: ffmpeg, go, air, swag, golangci-lint, migrate"
echo ""
print_info "開発サーバーを起動するには 'make dev' を実行してください"
