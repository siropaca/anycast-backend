-- Remove category_id from channels
DROP INDEX IF EXISTS idx_channels_category_id;
ALTER TABLE channels DROP COLUMN IF EXISTS category_id;

-- Drop categories table
DROP TABLE IF EXISTS categories;
