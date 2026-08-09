package server

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/openmtg/edh-go/pkg/ratelimit"
	"github.com/zeebo/errs"
)

// guestNameMaxAttempts bounds GuestSession's username_unique collision
// retry loop (T-02-04). Each attempt draws a fresh pairing (or, once the
// curated cross product is exhausted, a numeric-suffixed pairing per
// generateGuestName's overflow strategy) -- a visitor is never shown a
// naming failure (D-2.6/D-2.7).
const guestNameMaxAttempts = 8

// guestCredentialSecretBytes is the length, in raw bytes before base64url
// encoding, of the random secret half of the guest re-auth credential
// (DEC-D).
const guestCredentialSecretBytes = 32

// guestSessionOutcomeSuccess and guestSessionOutcomeFailure are the two
// bounded outcome values GuestSession ever passes to
// collectors.ObserveGuestSession -- never a caller-derived string, so the
// label stays bounded.
const (
	guestSessionOutcomeSuccess = "success"
	guestSessionOutcomeFailure = "failure"
)

// hashGuestSecret is the injectable seam GuestSession uses for both the
// guest's throwaway password and its re-auth credential hash. Production
// always leaves this at its zero value, hashPassword itself -- every real
// guest secret is bcrypt cost 14, identical to Signup's (T-02-05), never a
// cheaper hash for a "throwaway" account. It exists solely so
// TestGuestUsers_Create/GeneratedNameShape, which creates 50 guests in a
// tight loop to prove D-2.5/D-2.6's naming shape and uniqueness, can
// substitute a fast fake hash: under `go test -race`, a single bcrypt
// cost-14 hash measured over 8 seconds on this codebase's own dev hardware
// (versus well under a second without -race), and GuestSession performs two
// per call -- at real cost, 50 sequential creations alone would exceed
// `go test`'s default 10-minute timeout before any of this file's other
// tests could run. Swapping only the hash computation leaves the
// DB-enforced username_unique retry path -- the actual property this test
// proves -- completely real.
var hashGuestSecret = hashPassword

// guestCreationEnabled reports whether the guestSession mutation may run at
// all (DEC-A, T-02-02). Shaped like providerEnabled() (server/deck_providers.go),
// but with the opposite default: guest creation dials nobody, so the
// switch defaults true and exists to let an operator KILL the funnel, not
// to gate a rollout.
func (s *graphQLServer) guestCreationEnabled() bool {
	return s != nil && s.cfg.GuestCreationEnabled
}

// GuestSession creates a new immortal (D-2.1) guest identity: a users row
// with is_guest=true, expires_at left NULL (DEC-B), a generated
// MTG-flavored username (D-2.5/D-2.6) drawn from the same unique namespace
// as real usernames, an optional free-form non-unique display name
// (D-2.7/D-2.8), and a freshly minted guest re-auth credential (DEC-D). It
// returns a 24-hour JWT identical in shape to Signup/Login's, so a guest
// token satisfies requireAuth with no special-casing anywhere else in this
// codebase.
func (s *graphQLServer) GuestSession(ctx context.Context, displayName *string, sessionID string) (*User, error) {
	start := time.Now()
	outcome := guestSessionOutcomeFailure
	defer func() {
		collectors.ObserveGuestSession(outcome, time.Since(start))
	}()

	trimmedSession := strings.TrimSpace(sessionID)
	if trimmedSession == "" {
		return nil, errs.New("must provide a sessionID")
	}

	if !s.guestCreationEnabled() {
		return nil, errs.New("guest sessions are not available right now")
	}

	clientKey := clientKeyFor(ctx, trimmedSession)
	if !s.allowRequest(ctx, ratelimit.SurfaceGuestSession, clientKey) {
		return nil, errs.New("too many guest session requests -- please try again in a moment")
	}

	var trimmedDisplayName *string
	if displayName != nil {
		trimmed := strings.TrimSpace(*displayName)
		if trimmed != "" {
			trimmedDisplayName = &trimmed
		}
	}

	// The guest's password is crypto/rand and hashed through the same
	// bcrypt cost-14 path a real signup uses (T-02-05) -- never a cheaper
	// hash for a "throwaway" account. Nobody is ever told this value; the
	// only credential a guest client receives is the separate guest
	// credential minted below (DEC-D).
	rawPassword, err := randomBase64Secret(guestCredentialSecretBytes)
	if err != nil {
		return nil, errs.Wrap(err)
	}
	hashedPassword, err := hashGuestSecret(rawPassword)
	if err != nil {
		return nil, errs.Wrap(err)
	}

	id := uuid.New().String()

	credentialSecret, err := randomBase64Secret(guestCredentialSecretBytes)
	if err != nil {
		return nil, errs.Wrap(err)
	}
	credentialSecretHash, err := hashGuestSecret(credentialSecret)
	if err != nil {
		return nil, errs.Wrap(err)
	}
	// The credential returned to the client is the row's uuid, a single
	// "." separator, and the plaintext secret -- the uuid half is the O(1)
	// lookup key plan 02-02's re-auth path uses to find exactly one row
	// and bcrypt-compare once, instead of scanning every guest. Only the
	// secret half is ever hashed or stored.
	guestCredential := id + "." + credentialSecret

	stmt := `
	INSERT INTO "users" (uuid, username, password, is_guest, display_name, guest_credential_hash)
	VALUES ($1, $2, $3, true, $4, $5)
	RETURNING uuid, username, display_name, is_guest;
	`

	var user *User
	for attempt := 0; attempt < guestNameMaxAttempts; attempt++ {
		username, genErr := generateGuestName(attempt)
		if genErr != nil {
			return nil, errs.Wrap(genErr)
		}

		row, insertErr := s.insertGuestUser(ctx, stmt, id, username, hashedPassword, trimmedDisplayName, credentialSecretHash)
		if insertErr == nil {
			user = row
			break
		}
		if isUniqueViolation(insertErr) {
			// A generated name collided with an existing username (real or
			// guest). Redraw and retry -- a visitor must never see a naming
			// failure (D-2.6/D-2.7's prohibition).
			continue
		}
		return nil, errs.Wrap(insertErr)
	}
	if user == nil {
		return nil, errs.New("failed to allocate a unique guest name")
	}

	token, err := newAuthToken(user, time.Hour*24)
	if err != nil {
		return nil, errs.Wrap(err)
	}
	user.Token = &token
	user.GuestCredential = &guestCredential
	isGuest := true
	user.IsGuest = &isGuest

	userID := user.ID
	s.recordProductEvent(ctx, ProductEvent{
		Name:      "guest_session_created",
		SessionID: trimmedSession,
		UserID:    &userID,
	}, false)

	outcome = guestSessionOutcomeSuccess
	return user, nil
}

// insertGuestUser runs the single-row INSERT and scans its RETURNING
// clause into a *User. Split out of GuestSession so the collision-retry
// loop above stays readable.
func (s *graphQLServer) insertGuestUser(ctx context.Context, stmt string, id, username, hashedPassword string, displayName *string, credentialHash string) (*User, error) {
	rows, err := s.db.QueryContext(ctx, stmt, id, username, hashedPassword, displayName, credentialHash)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	user := &User{}
	for rows.Next() {
		if scanErr := rows.Scan(&user.ID, &user.Username, &user.DisplayName, &user.IsGuest); scanErr != nil {
			return nil, scanErr
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return user, nil
}

// isUniqueViolation reports whether err represents a username_unique (or
// equivalent) constraint violation, mirroring Signup's own substring check
// (server/users.go) rather than depending on a specific driver error type.
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "username_unique") || strings.Contains(msg, "duplicate key value")
}

// randomBase64Secret returns n raw random bytes from crypto/rand, encoded
// base64url without padding.
func randomBase64Secret(n int) (string, error) {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("guest session: generate random secret: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
