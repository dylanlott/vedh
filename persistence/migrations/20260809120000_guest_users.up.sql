-- Guest identity columns (REQ-ACT-005, D-2.1..D-2.8). is_guest distinguishes
-- an immortal guest row (D-2.1: guests never expire, expires_at stays NULL
-- for a live guest per DEC-B) from a real signup. expires_at stays on the
-- schema and stays enforced in authorization (D-2.2) but governs only
-- whether a session credential may mint a new token -- it never triggers
-- row deletion. display_name is deliberately NOT unique (D-2.7/D-2.8): the
-- unique username_unique constraint stays on username alone, and a visitor
-- typed display name may collide with any other user's without ever
-- surfacing a "taken" error. guest_credential_hash stores only the bcrypt
-- hash of the guest re-auth credential (DEC-D) -- the plaintext credential
-- is returned to the client exactly once and never persisted.
ALTER TABLE users ADD COLUMN is_guest BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE users ADD COLUMN expires_at TIMESTAMPTZ;
ALTER TABLE users ADD COLUMN display_name VARCHAR(255);
ALTER TABLE users ADD COLUMN guest_credential_hash VARCHAR(255);
