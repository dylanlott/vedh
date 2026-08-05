-- True inverse of 20260804120100_card_name_search.up.sql, in reverse
-- statement order. Deliberately does NOT drop the pg_trgm extension: it is
-- a shared, database-wide resource that a later migration or another
-- feature may depend on, and dropping it here would be a footgun.
DROP INDEX IF EXISTS cards_name_set_num_idx;
DROP INDEX IF EXISTS cards_facename_lower_idx;
DROP INDEX IF EXISTS cards_name_lower_idx;
DROP INDEX IF EXISTS card_names_trgm_gist;
DROP TABLE IF EXISTS card_names;
