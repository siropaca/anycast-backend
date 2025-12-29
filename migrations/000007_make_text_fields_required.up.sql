-- description と persona を NOT NULL に変更

-- channels.description
UPDATE channels SET description = '' WHERE description IS NULL;
ALTER TABLE channels ALTER COLUMN description SET NOT NULL;

-- characters.persona
UPDATE characters SET persona = '' WHERE persona IS NULL;
ALTER TABLE characters ALTER COLUMN persona SET NOT NULL;

-- episodes.description
UPDATE episodes SET description = '' WHERE description IS NULL;
ALTER TABLE episodes ALTER COLUMN description SET NOT NULL;

-- sound_effects.description
UPDATE sound_effects SET description = '' WHERE description IS NULL;
ALTER TABLE sound_effects ALTER COLUMN description SET NOT NULL;
