-- ============================================
-- Remove is_active from categories
-- ============================================

DROP INDEX IF EXISTS idx_categories_is_active;

ALTER TABLE categories
DROP COLUMN is_active;
