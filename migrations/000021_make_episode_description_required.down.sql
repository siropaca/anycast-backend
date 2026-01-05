-- episodes.description を NULL 許可に戻す
ALTER TABLE episodes ALTER COLUMN description DROP NOT NULL;
