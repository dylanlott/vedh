ALTER TABLE IF EXISTS gamelog
    ADD COLUMN IF NOT EXISTS id BIGSERIAL,
    ADD COLUMN IF NOT EXISTS game_id TEXT;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM   information_schema.table_constraints tc
        WHERE  tc.table_name = 'gamelog'
        AND    tc.constraint_type = 'PRIMARY KEY'
    ) THEN
        ALTER TABLE gamelog ADD PRIMARY KEY (id);
    END IF;
END$$;

UPDATE gamelog SET game_id = '' WHERE game_id IS NULL;

ALTER TABLE IF EXISTS gamelog
    ALTER COLUMN game_id SET NOT NULL,
    ALTER COLUMN eventtime SET DEFAULT now();
