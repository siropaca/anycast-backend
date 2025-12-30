-- channels テーブルに台本生成プロンプトカラムを追加
ALTER TABLE channels ADD COLUMN script_prompt TEXT;

-- episodes テーブルに台本生成プロンプトカラムを追加
ALTER TABLE episodes ADD COLUMN script_prompt TEXT;
