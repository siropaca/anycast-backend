-- ============================================
-- Fix script_lines foreign key constraints
-- ============================================
-- Change speaker_id and sfx_id from ON DELETE RESTRICT to ON DELETE CASCADE
-- This allows channels to be deleted properly (cascades through characters and sound_effects)

-- Drop existing constraints
ALTER TABLE script_lines DROP CONSTRAINT script_lines_speaker_id_fkey;
ALTER TABLE script_lines DROP CONSTRAINT script_lines_sfx_id_fkey;

-- Re-create with ON DELETE CASCADE
ALTER TABLE script_lines
ADD CONSTRAINT script_lines_speaker_id_fkey
FOREIGN KEY (speaker_id) REFERENCES characters(id) ON DELETE CASCADE;

ALTER TABLE script_lines
ADD CONSTRAINT script_lines_sfx_id_fkey
FOREIGN KEY (sfx_id) REFERENCES sound_effects(id) ON DELETE CASCADE;
