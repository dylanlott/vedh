-- pg_trgm ships with PostgreSQL 14 marked trusted, so an ordinary database
-- owner with CREATE privilege can install it; this migration never requests
-- and never requires a superuser step.
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- card_names is a projection over cards: one row per distinct lower-cased
-- searchable name across the name and face-name columns, not one row per
-- printing. The cards table holds roughly ninety-seven thousand printings
-- against roughly thirty-three thousand distinct names, so indexing
-- printings directly would produce a three-times larger index whose top
-- three results are frequently three printings of the same card -- useless
-- as three suggestions.
CREATE TABLE IF NOT EXISTS card_names (
    name_lower TEXT PRIMARY KEY,
    display    TEXT NOT NULL
);

INSERT INTO card_names (name_lower, display)
SELECT DISTINCT lower(n.the_name) AS name_lower, n.the_name AS display
FROM (
    SELECT name AS the_name FROM cards WHERE name IS NOT NULL AND name <> ''
    UNION ALL
    SELECT facename AS the_name FROM cards WHERE facename IS NOT NULL AND facename <> ''
) AS n
ON CONFLICT (name_lower) DO NOTHING;

-- GiST rather than GIN is decided by correctness, not benchmarking: the GIN
-- trigram operator class only accelerates the boolean similarity operator
-- `%`, which returns zero rows for a needle below the similarity
-- threshold -- but D-03 requires the nearest matches to be shown even below
-- the cutoff. The GiST distance operator `<->` answers that in one
-- index-ordered KNN scan, unconditionally.
CREATE INDEX IF NOT EXISTS card_names_trgm_gist
    ON card_names USING GIST (name_lower gist_trgm_ops);

-- Expression indexes on cards. An expression index over a lower-cased
-- column is unusable by a query that does not also lower-case its
-- comparison, so server/cards.go's WHERE clause changes alongside this
-- migration.
CREATE INDEX IF NOT EXISTS cards_name_lower_idx
    ON cards (lower(name));

CREATE INDEX IF NOT EXISTS cards_facename_lower_idx
    ON cards (lower(facename));

-- D-09 printing disambiguation: (name, set code, collector number). Set
-- code is lower-cased in the index (and at query time) so "C21" and "c21"
-- select the same row; collector number is compared as stored.
CREATE INDEX IF NOT EXISTS cards_name_set_num_idx
    ON cards (lower(name), lower(setcode), number);
