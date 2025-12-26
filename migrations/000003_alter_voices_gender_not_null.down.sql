-- gender を NULL 許容に戻す
ALTER TABLE voices ALTER COLUMN gender DROP NOT NULL;
