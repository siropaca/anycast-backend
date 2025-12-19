.PHONY: dev run build test fmt lint tidy clean

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
