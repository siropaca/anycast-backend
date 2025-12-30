-- episodes テーブルから台本生成プロンプトカラムを削除
ALTER TABLE episodes DROP COLUMN IF EXISTS script_prompt;

-- channels テーブルから台本生成プロンプトカラムを削除
ALTER TABLE channels DROP COLUMN IF EXISTS script_prompt;
