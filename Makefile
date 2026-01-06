.PHONY: bootstrap bs dev run build test fmt lint lint-fix tidy clean swagger migrate-up migrate-down migrate-reset seed token cleanup cleanup-run

DATABASE_URL ?= postgres://postgres:postgres@localhost:5433/anycast?sslmode=disable

# 開発ツールをインストール
bootstrap:
	./scripts/bootstrap.sh

bs: bootstrap

# 開発サーバーを起動（ホットリロード）
dev:
	mise exec -- air

# サーバーを起動
run:
	go run main.go

# バイナリをビルド
build:
	go build -o bin/server main.go

# テストを実行
test:
	go test ./...

# コードをフォーマット
fmt:
	gofmt -w .

# 静的解析を実行
lint:
	mise exec -- golangci-lint run ./...

# 静的解析を実行（自動修正あり）
lint-fix:
	mise exec -- golangci-lint run --fix ./...

# 依存関係を整理
tidy:
	go mod tidy

# ビルド成果物を削除
clean:
	rm -rf bin/ tmp/

# Swagger ドキュメント生成
swagger:
	mise exec -- swag init -g main.go -o swagger --outputTypes go,json

# マイグレーション実行
migrate-up:
	mise exec -- migrate -path migrations -database "$(DATABASE_URL)" up

# マイグレーションロールバック
migrate-down:
	mise exec -- migrate -path migrations -database "$(DATABASE_URL)" down

# マイグレーションリセット（テーブル全削除 → 再マイグレーション）
migrate-reset:
	docker exec -i anycast-db psql -U postgres -d anycast -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"
	mise exec -- migrate -path migrations -database "$(DATABASE_URL)" up

# シードデータを投入（開発環境用）
seed:
	@for file in seeds/*.sql; do \
		echo "Running $$file..."; \
		docker exec -i anycast-db psql -U postgres -d anycast < "$$file"; \
	done

# 開発用 JWT トークンを生成
token:
	@go run ./scripts/gentoken

# 孤児メディアファイルをクリーンアップ（dry-run）
cleanup:
	@go run ./scripts/cleanup --dry-run=true

# 孤児メディアファイルをクリーンアップ（実行）
cleanup-run:
	@go run ./scripts/cleanup --dry-run=false
