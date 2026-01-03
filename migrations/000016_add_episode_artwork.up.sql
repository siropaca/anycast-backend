-- episodes テーブルに artwork_id カラムを追加
ALTER TABLE episodes ADD COLUMN artwork_id UUID REFERENCES images(id);
