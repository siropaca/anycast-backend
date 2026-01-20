# ADR-015: ユーザー権限管理に Enum 型を使用

## ステータス

Accepted

## コンテキスト

サービスの運用にあたり、ユーザーに権限（ロール）を持たせる必要が生じた。具体的には以下の要件がある：

- 一般ユーザー（user）: 自分のコンテンツのみ管理可能
- 管理者（admin）: マスタデータ（Voice, Category）の CRUD、全ユーザーのコンテンツ閲覧・編集・削除が可能

権限管理の設計方針を決定する必要がある。

## 決定

PostgreSQL の Enum 型（`user_role`）を使用し、users テーブルに `role` カラムを追加する。

## 選択肢

### 選択肢 1: Enum 型（user_role）

users テーブルに `role` カラム（Enum 型）を追加する。

```sql
CREATE TYPE user_role AS ENUM ('user', 'admin');
ALTER TABLE users ADD COLUMN role user_role NOT NULL DEFAULT 'user';
```

- メリット
  - シンプルで実装が容易
  - DB レベルで値の制約を保証できる
  - 追加のテーブル結合が不要でパフォーマンスに優れる
- デメリット
  - 新しいロールの追加には Enum 型の変更が必要
  - 細かい権限制御（パーミッション単位）には向かない

### 選択肢 2: ロールテーブル分離

roles テーブルを作成し、users テーブルから外部キーで参照する。

```sql
CREATE TABLE roles (
    id UUID PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE,
    description TEXT
);
ALTER TABLE users ADD COLUMN role_id UUID REFERENCES roles(id);
```

- メリット
  - 新しいロールの追加が容易（INSERT のみ）
  - ロールにメタデータ（説明など）を持たせられる
  - 将来的に多対多（ユーザーが複数ロールを持つ）への拡張が容易
- デメリット
  - 実装が複雑になる
  - テーブル結合が必要になる
  - 現時点では過剰な設計

## 理由

現時点で必要なロールは `user` と `admin` の 2 種類のみであり、今後も大幅に増える予定はない。シンプルな要件に対して複雑な設計を採用することは、メンテナンスコストの増加につながる。

Enum 型であれば DB レベルで値の整合性を保証でき、Go コード側でも定数として扱えるため、型安全性も確保できる。将来的に複雑な権限管理が必要になった場合は、その時点で ADR を更新してロールテーブル方式に移行すればよい。

## 結果

- users テーブルに `role` カラム（user_role Enum 型）を追加
- User モデルに `Role` フィールドを追加
- Admin 権限チェックミドルウェアを実装し、管理者専用 API に適用
- 既存ユーザーはすべて `user` ロールとして扱う（デフォルト値）
