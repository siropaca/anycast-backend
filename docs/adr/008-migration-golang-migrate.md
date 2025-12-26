# ADR-008: マイグレーションツールとして golang-migrate を採用

## ステータス

Accepted

## コンテキスト

データベーススキーマのバージョン管理を行う方法を選定する必要があった。以下の要件を考慮した:

- SQL ベースのマイグレーション
- Up/Down のサポート
- CI/CD との統合
- 複数環境での実行

## 決定

**golang-migrate** を採用する。

```bash
# マイグレーション実行
migrate -path migrations -database "$DATABASE_URL" up

# ロールバック
migrate -path migrations -database "$DATABASE_URL" down 1
```

## 選択肢

### 選択肢 1: golang-migrate

- 純粋な SQL ファイルでマイグレーション定義
- CLI ツールとして使用可能
- Docker イメージあり
- 多くのデータベースをサポート
- Railway などの PaaS との統合が容易

### 選択肢 2: GORM AutoMigrate

- コードからスキーマを自動生成
- 開発時は便利
- Down マイグレーションがない
- 本番環境での使用は非推奨

### 選択肢 3: goose

- Go で書かれたマイグレーションツール
- SQL と Go 両方でマイグレーション可能
- golang-migrate より機能は少ない

### 選択肢 4: Atlas

- 宣言的スキーマ管理
- スキーマ差分から自動でマイグレーション生成
- 学習コストが高い
- 新しいツールで実績が少ない

### 選択肢 5: sql-migrate

- SQL ベースのマイグレーション
- Go からも CLI からも使用可能
- メンテナンス頻度が低い

## 理由

1. **シンプルさ**: 純粋な SQL ファイルで管理
2. **CLI ツール**: mise でインストールして使用
3. **CI/CD 統合**: Makefile から簡単に実行可能
4. **Railway 対応**: 起動時に自動マイグレーション実行可能
5. **広く使われている**: 情報が豊富

## 結果

- `migrations/` ディレクトリにマイグレーションファイルを配置
- ファイル名は `XXXXXX_description.up.sql` / `XXXXXX_description.down.sql`
- Makefile に `migrate-up`, `migrate-down`, `migrate-status` を定義
- Railway では起動コマンドでマイグレーション実行
