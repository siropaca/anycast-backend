-- description と persona を NULLABLE に戻す

ALTER TABLE channels ALTER COLUMN description DROP NOT NULL;
ALTER TABLE characters ALTER COLUMN persona DROP NOT NULL;
ALTER TABLE episodes ALTER COLUMN description DROP NOT NULL;
ALTER TABLE sound_effects ALTER COLUMN description DROP NOT NULL;
