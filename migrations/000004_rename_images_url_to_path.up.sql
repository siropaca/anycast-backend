-- images テーブルの url カラムを path にリネーム
ALTER TABLE images RENAME COLUMN url TO path;

-- カラムコメントを更新（オプション）
COMMENT ON COLUMN images.path IS 'GCS 上のパス（例: images/xxx.png）';
