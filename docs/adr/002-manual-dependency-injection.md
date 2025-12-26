# ADR-002: 手動による依存性注入

## ステータス

Accepted

## コンテキスト

レイヤードアーキテクチャを採用するにあたり、各レイヤー間の依存関係を管理する方法を決定する必要があった。  
テスタビリティを確保するため、依存性注入（DI）パターンの採用は必須と判断した。

## 決定

**手動 DI** を採用する。

`internal/di/container.go` に DI コンテナを実装し、`main.go` で依存関係を構築する。

```go
type Container struct {
    VoiceHandler *handler.VoiceHandler
}

func NewContainer(db *gorm.DB) *Container {
    voiceRepo := repository.NewVoiceRepository(db)
    voiceService := service.NewVoiceService(voiceRepo)
    voiceHandler := handler.NewVoiceHandler(voiceService)
    return &Container{VoiceHandler: voiceHandler}
}
```

## 選択肢

### 選択肢 1: 手動 DI

- シンプルで理解しやすい
- 追加の依存関係が不要
- コンパイル時に依存関係エラーを検出可能
- 依存関係が増えると手動管理が煩雑になる可能性

### 選択肢 2: Google Wire

- コード生成による型安全な DI
- ボイラープレート削減
- 学習コストがある
- 大規模プロジェクト向け

### 選択肢 3: Uber fx

- ランタイム DI
- 自動的な依存解決
- リフレクションベースでデバッグが困難
- 依存関係エラーが実行時まで検出されない

### 選択肢 4: グローバル変数

- 最もシンプル
- テストが困難
- 暗黙的な依存関係

## 理由

1. **シンプルさ**: 現時点の依存関係数では手動管理で十分
2. **透明性**: 依存関係が明示的でコードを読めば理解できる
3. **段階的移行**: 将来的に Wire への移行が容易
4. **外部依存の最小化**: 追加ライブラリ不要

## 結果

- `internal/di/container.go` で全ての依存関係を構築
- 新しいハンドラー追加時は `Container` struct と `NewContainer` 関数を更新
- 依存関係が 20+ を超えた場合は Wire への移行を検討
