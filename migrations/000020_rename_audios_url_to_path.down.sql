-- audios テーブルの path カラムを url に戻す
ALTER TABLE audios RENAME COLUMN path TO url;
