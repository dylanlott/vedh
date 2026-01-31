DELETE FROM allcards WHERE uuid LIKE 'testseed-%';
DELETE FROM cards WHERE uuid LIKE 'testseed-%';
DROP INDEX IF EXISTS allcards_facename_idx;
DROP INDEX IF EXISTS allcards_name_idx;
DROP TABLE IF EXISTS allcards;
DROP TABLE IF EXISTS cards CASCADE;
