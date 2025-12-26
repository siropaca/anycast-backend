-- gender が NULL のレコードにデフォルト値を設定
UPDATE voices SET gender = 'neutral' WHERE gender IS NULL;

-- gender を NOT NULL に変更
ALTER TABLE voices ALTER COLUMN gender SET NOT NULL;
