# ADR-011: テストライブラリ: testify

## ステータス

Accepted

## コンテキスト

Go のユニットテストを作成するにあたり、アサーションの記述方法を検討する必要があった。  
標準ライブラリの `testing` パッケージだけでもテストは書けるが、アサーションの記述が冗長になりがちである。

## 決定

テストライブラリとして [stretchr/testify](https://github.com/stretchr/testify) を採用する。

## 選択肢

### 選択肢 1: 標準ライブラリのみ

```go
if got != want {
    t.Errorf("got %v, want %v", got, want)
}
```

- メリット
  - 外部依存がない
  - Go の標準的な書き方を学べる
- デメリット
  - アサーションが冗長になる
  - エラーメッセージを毎回書く必要がある

### 選択肢 2: testify

```go
assert.Equal(t, want, got)
```

- メリット
  - アサーションが簡潔に書ける
  - 豊富なアサーション関数（`Equal`, `Nil`, `ErrorIs`, `Len` など）
  - モック機能（`mock` パッケージ）も提供
  - Go コミュニティで最も広く使われている
- デメリット
  - 外部依存が増える

### 選択肢 3: go-cmp

```go
if diff := cmp.Diff(want, got); diff != "" {
    t.Errorf("mismatch (-want +got):\n%s", diff)
}
```

- メリット
  - 構造体の差分を見やすく表示
  - Google 製
- デメリット
  - アサーション機能は限定的
  - testify ほど多機能ではない

## 理由

1. **簡潔さ**: testify を使うことでテストコードが読みやすくなる
2. **普及度**: Go のテストライブラリとして最も人気があり、情報が豊富
3. **機能**: `assert` / `require` / `mock` など必要な機能が揃っている
4. **将来性**: モックが必要になった際も同じライブラリで対応できる

## 結果

- `github.com/stretchr/testify` を依存関係に追加
- テストでは `assert` パッケージを使用してアサーションを記述
- 失敗時に即座に終了したい場合は `require` パッケージを使用
