-- ============================================
-- Add is_active to categories
-- ============================================

ALTER TABLE categories
ADD COLUMN is_active BOOLEAN NOT NULL DEFAULT true;

CREATE INDEX idx_categories_is_active ON categories (is_active);
