# Definition of Done（完了の定義）

タスクが「完了」と見なされるために満たすべき基準。PR 作成時にチェックリストとして使用する。

---

## コード品質

- [ ] `make fmt` でフォーマット済み
- [ ] `make lint` でエラーなし
- [ ] `make test` で全テスト通過
- [ ] 既存のコーディング規約に準拠（[conventions.md](conventions.md)）

## テスト

- [ ] 外部依存のないユニットテストを実装した
- [ ] http/ 配下の .http ファイルで手動テストを確認した（API の場合）

## ドキュメント

以下は該当する場合のみ。

- [ ] `make swagger` で Swagger ドキュメントを再生成した
- [ ] docs/api/INDEX.md の API 一覧テーブルを更新した
- [ ] ドメインモデル（docs/domain-model/）を更新した
- [ ] DB 設計（docs/specs/database.md）を更新した
- [ ] ADR を作成し、docs/adr/INDEX.md の一覧に追記した
- [ ] .env.example を更新した
- [ ] .http ファイルを作成・更新した
