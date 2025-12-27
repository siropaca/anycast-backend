# ADR-010: 静的解析ツールとして golangci-lint を採用

## ステータス

Accepted

## コンテキスト

コード品質を維持し、バグを早期に発見するために静的解析ツールを導入する必要があった。  
Go 標準の `go vet` だけでは検出できる問題が限られていた。

## 決定

**golangci-lint** を採用する。

```bash
# チェックのみ
make lint

# 自動修正あり
make lint-fix
```

有効にした linter:

| linter | 検出内容 |
|--------|----------|
| govet | 怪しいコードパターン |
| errcheck | エラーの無視 |
| staticcheck | バグになりやすいコード |
| unused | 使われていないコード |
| ineffassign | 無駄な代入 |
| gocritic | コード改善提案 |
| nilnil | (nil, nil) を返していないか |
| gofmt / goimports | フォーマット |

## 選択肢

### 選択肢 1: golangci-lint

- **メリット**
  - 複数の linter を統合したオールインワンツール
  - 設定ファイル（`.golangci.yml`）で一元管理
  - CI/CD との統合が容易
  - Go 界隈でデファクトスタンダード
- **デメリット**
  - 特になし

### 選択肢 2: go vet のみ

- **メリット**
  - Go 標準ツール
  - 追加インストール不要
- **デメリット**
  - 検出できる問題が限られる

### 選択肢 3: 個別の linter を組み合わせ

- **メリット**
  - 必要なものだけ選択可能
- **デメリット**
  - 管理が煩雑
  - 実行コマンドが複数必要

### 選択肢 4: staticcheck 単体

- **メリット**
  - 高品質な単体 linter
  - golangci-lint より軽量
- **デメリット**
  - 他の linter と組み合わせが必要

## 理由

1. **包括性**: 複数の linter を一つのツールで実行できる
2. **設定の一元管理**: `.golangci.yml` で全ての設定を管理
3. **エコシステム**: 多くの Go プロジェクトで採用されている
4. **拡張性**: 必要に応じて linter の追加・削除が容易

## 結果

- `.mise.toml` で golangci-lint をインストール
- `.golangci.yml` で linter の設定を管理
- `make lint` / `make lint-fix` で実行
- コミット前に `make lint` を実行することを推奨
