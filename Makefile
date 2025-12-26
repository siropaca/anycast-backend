.PHONY: dev run build test fmt lint tidy clean migrate-up migrate-down migrate-reset migrate-status swagger

DATABASE_URL ?= postgres://postgres:postgres@localhost:5432/anycast?sslmode=disable

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
	go vet ./...

# 依存関係を整理
tidy:
	go mod tidy

# ビルド成果物を削除
clean:
	rm -rf bin/ tmp/

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

# マイグレーション状態確認
migrate-status:
	migrate -path migrations -database "$(DATABASE_URL)" version

# Swagger ドキュメント生成
swagger:
	swag init -g main.go -o swagger
