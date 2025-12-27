.PHONY: dev run build test fmt lint lint-fix tidy clean migrate-up migrate-down migrate-reset swagger bootstrap bs seed token

DATABASE_URL ?= postgres://postgres:postgres@localhost:5433/anycast?sslmode=disable

# 開発ツールをインストール
bootstrap:
	./scripts/bootstrap.sh

bs: bootstrap

# 開発サーバーを起動（ホットリロード）
dev:
	air

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
	golangci-lint run ./...

# 静的解析を実行（自動修正あり）
lint-fix:
	golangci-lint run --fix ./...

# 依存関係を整理
tidy:
	go mod tidy

# ビルド成果物を削除
clean:
	rm -rf bin/ tmp/

# Swagger ドキュメント生成
swagger:
	swag init -g main.go -o swagger

# マイグレーション実行
migrate-up:
	migrate -path migrations -database "$(DATABASE_URL)" up

# マイグレーションロールバック
migrate-down:
	migrate -path migrations -database "$(DATABASE_URL)" down

# マイグレーションリセット（down → up）
migrate-reset:
	migrate -path migrations -database "$(DATABASE_URL)" down -all
	migrate -path migrations -database "$(DATABASE_URL)" up

# シードデータを投入（開発環境用）
seed:
	@for file in seeds/*.sql; do \
		echo "Running $$file..."; \
		docker exec -i anycast-db psql -U postgres -d anycast < "$$file"; \
	done

# 開発用 JWT トークンを生成
token:
	@go run ./scripts/gentoken
