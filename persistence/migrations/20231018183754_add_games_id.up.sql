ALTER TABLE games ADD COLUMN id text;
CREATE INDEX idx_games_id ON games(id);
