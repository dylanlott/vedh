package server

import "context"

// CleanupExpiredGuests deletes only guest rows an operator explicitly marked
// with a past expires_at and which no games payload mentions by UUID or
// username. D-2.1 creates guests with expires_at NULL, so this is an
// independently retryable safety valve with no scheduled production work; no
// code in this package may schedule it as a live-guest reaper.
func (s *graphQLServer) CleanupExpiredGuests(ctx context.Context) (int64, error) {
	result, err := s.db.ExecContext(ctx, `
		DELETE FROM users AS u
		WHERE u.is_guest = true
		  AND u.expires_at IS NOT NULL
		  AND u.expires_at <= CURRENT_TIMESTAMP
		  AND NOT EXISTS (
			SELECT 1
			FROM games AS g
			WHERE strpos(COALESCE(g.payload::text, ''), u.uuid) > 0
			   OR strpos(COALESCE(g.payload::text, ''), u.username) > 0
		  )
	`)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
