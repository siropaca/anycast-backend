-- images テーブルの path カラムを url に戻す
ALTER TABLE images RENAME COLUMN path TO url;

-- カラムコメントを削除
COMMENT ON COLUMN images.url IS NULL;
