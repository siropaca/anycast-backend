-- episodes.description を NOT NULL に変更
UPDATE episodes SET description = '' WHERE description IS NULL;
ALTER TABLE episodes ALTER COLUMN description SET NOT NULL;
