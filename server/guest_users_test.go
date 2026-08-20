package server

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/openmtg/edh-go/pkg/ratelimit"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"golang.org/x/crypto/bcrypt"
)

// guestTestSessionID returns a session ID unique to the calling (sub)test,
// mirroring server/product_events_test.go's uniqueSessionID.
func guestTestSessionID(t *testing.T) string {
	t.Helper()
	return t.Name() + "-" + strconv.FormatInt(time.Now().UnixNano(), 10)
}

func countUsers(t *testing.T, s *graphQLServer) int {
	t.Helper()
	var count int
	if err := s.db.QueryRow(`SELECT count(*) FROM users`).Scan(&count); err != nil {
		t.Fatalf("count users: %v", err)
	}
	return count
}

// TestGuestHost_Tracer is the architectural proof this whole plan exists to
// deliver: a guest JWT satisfies requireAuth with no special-casing
// anywhere else in the codebase, all the way through the existing
// CreateGame.
func TestGuestHost_Tracer(t *testing.T) {
	s := testAPI(t)
	sessionID := guestTestSessionID(t)

	user, err := s.GuestSession(context.Background(), nil, sessionID)
	if err != nil {
		t.Fatalf("GuestSession() error = %v", err)
	}
	if user == nil {
		t.Fatal("GuestSession() returned a nil user")
	}
	if user.IsGuest == nil || !*user.IsGuest {
		t.Fatal("expected IsGuest = true")
	}
	if user.Token == nil || *user.Token == "" {
		t.Fatal("expected a non-empty Token")
	}
	if user.GuestCredential == nil || *user.GuestCredential == "" {
		t.Fatal("expected a non-empty GuestCredential")
	}
	if user.Password != nil {
		t.Fatal("expected Password to be nil on the returned user")
	}

	// Exactly one users row, is_guest true, expires_at NULL, a non-empty
	// bcrypt password distinct from the returned credential.
	var (
		rowCount            int
		password            string
		isGuest             bool
		expiresAtIsNull     bool
		guestCredentialHash string
	)
	if err := s.db.QueryRow(
		`SELECT count(*) FROM users WHERE uuid = $1`, user.ID,
	).Scan(&rowCount); err != nil {
		t.Fatalf("count guest rows: %v", err)
	}
	if rowCount != 1 {
		t.Fatalf("expected exactly one users row for the new guest, got %d", rowCount)
	}
	if err := s.db.QueryRow(
		`SELECT password, is_guest, (expires_at IS NULL), guest_credential_hash FROM users WHERE uuid = $1`, user.ID,
	).Scan(&password, &isGuest, &expiresAtIsNull, &guestCredentialHash); err != nil {
		t.Fatalf("scan guest row: %v", err)
	}
	if password == "" {
		t.Fatal("expected a non-empty bcrypt password")
	}
	if password == *user.GuestCredential {
		t.Fatal("expected the stored password to differ from the returned guest credential")
	}
	if !isGuest {
		t.Fatal("expected is_guest = true in the database")
	}
	if !expiresAtIsNull {
		t.Fatal("expected expires_at IS NULL (DEC-B: guest rows are immortal)")
	}
	if guestCredentialHash == "" {
		t.Fatal("expected a non-empty guest_credential_hash")
	}

	// Parse the returned Token exactly as an ordinary HTTP request would,
	// then use it to call the existing CreateGame with no special-casing.
	authUser, err := parseAndValidateToken(*user.Token)
	if err != nil {
		t.Fatalf("parseAndValidateToken() error = %v", err)
	}
	ctx := withAuth(context.Background(), authUser)

	gameID := "guest-tracer-" + sessionID
	decklist := "1,Sol Ring"
	game, err := s.CreateGame(ctx, InputCreateGame{
		ID:   gameID,
		Turn: &InputTurn{Player: authUser.Username, Phase: "MAIN", Number: 1, Priority: authUser.Username},
		Players: []*InputBoardState{
			{
				UserID:   authUser.ID,
				User:     authUser.Username,
				GameID:   gameID,
				Life:     40,
				Decklist: &decklist,
			},
		},
	})
	t.Cleanup(func() {
		_, _ = s.db.Exec(`DELETE FROM games WHERE id = $1`, gameID)
	})
	if err != nil {
		t.Fatalf("CreateGame() with a guest token error = %v", err)
	}
	if game == nil || game.ID != gameID {
		t.Fatalf("expected a created game with ID %q, got %+v", gameID, game)
	}
}

// TestGuestUsers_Create groups the guest-creation shape assertions D-2.5
// through D-2.8 require.
func TestGuestUsers_Create(t *testing.T) {
	t.Run("GeneratedNameShape", func(t *testing.T) {
		s := testAPI(t)

		// This subtest creates 50 guests in a tight loop to prove
		// D-2.5/D-2.6's naming shape and uniqueness across many sequential
		// creations. hashGuestSecret is swapped for a fast fake so the
		// wall-clock cost of GuestSession's two real bcrypt cost-14 hashes
		// per call -- roughly 8s each under `go test -race` on this
		// codebase's own dev hardware -- does not push this subtest alone
		// past `go test`'s default 10-minute timeout. The DB-enforced
		// username_unique retry path this subtest actually proves is left
		// completely real; only the hash computation is faked.
		originalHash := hashGuestSecret
		hashGuestSecret = func(secret string) (string, error) { return "fake-hash:" + secret, nil }
		t.Cleanup(func() { hashGuestSecret = originalHash })

		seen := map[string]bool{}
		for i := 0; i < 50; i++ {
			sessionID := guestTestSessionID(t) + "-" + strconv.Itoa(i)
			user, err := s.GuestSession(context.Background(), nil, sessionID)
			if err != nil {
				t.Fatalf("GuestSession() iteration %d error = %v", i, err)
			}
			parts := strings.Split(user.Username, " ")
			if len(parts) < 2 {
				t.Fatalf("expected an adjective-noun pairing, got %q", user.Username)
			}
			if !isGuestAdjective(parts[0]) {
				t.Errorf("username %q: %q is not a curated adjective", user.Username, parts[0])
			}
			if !isGuestNoun(parts[1]) {
				t.Errorf("username %q: %q is not a curated noun", user.Username, parts[1])
			}
			if seen[user.Username] {
				t.Fatalf("username %q was generated twice across 50 sequential creations", user.Username)
			}
			seen[user.Username] = true
		}
	})

	t.Run("DisplayNameStored", func(t *testing.T) {
		s := testAPI(t)
		typedName := "Dylan"

		first, err := s.GuestSession(context.Background(), &typedName, guestTestSessionID(t)+"-a")
		if err != nil {
			t.Fatalf("GuestSession() (first) error = %v", err)
		}
		if first.DisplayName == nil || *first.DisplayName != typedName {
			t.Fatalf("expected DisplayName = %q, got %+v", typedName, first.DisplayName)
		}

		second, err := s.GuestSession(context.Background(), &typedName, guestTestSessionID(t)+"-b")
		if err != nil {
			t.Fatalf("GuestSession() (second) error = %v", err)
		}
		if second.DisplayName == nil || *second.DisplayName != typedName {
			t.Fatalf("expected DisplayName = %q, got %+v", typedName, second.DisplayName)
		}
		if second.Username == first.Username {
			t.Fatal("expected the two guests to receive distinct generated usernames despite the same typed display name")
		}
	})

	t.Run("CollisionRetries", func(t *testing.T) {
		s := testAPI(t)

		collidingUsername := "Brave Sliver"
		if _, err := s.Signup(context.Background(), collidingUsername, "seedpassword123"); err != nil {
			t.Fatalf("failed to seed the colliding username: %v", err)
		}

		// Force generateGuestName's first draw to the pre-taken pairing,
		// then let subsequent draws (the retry) proceed normally.
		originalDraw := drawGuestNamePair
		calls := 0
		drawGuestNamePair = func() (string, string, error) {
			calls++
			if calls == 1 {
				return "Brave", "Sliver", nil
			}
			return originalDraw()
		}
		t.Cleanup(func() { drawGuestNamePair = originalDraw })

		user, err := s.GuestSession(context.Background(), nil, guestTestSessionID(t))
		if err != nil {
			t.Fatalf("GuestSession() error = %v, want a successful retry past the collision", err)
		}
		if user.Username == collidingUsername {
			t.Fatalf("expected a different username than the pre-taken %q", collidingUsername)
		}
		if calls < 2 {
			t.Fatalf("expected at least 2 draw attempts (the forced collision plus a retry), got %d", calls)
		}
	})
}

// TestGuestUsers_KillSwitch proves T-02-02: with the kill switch off,
// GuestSession refuses to run and writes zero rows.
func TestGuestUsers_KillSwitch(t *testing.T) {
	s := testAPI(t)
	s.cfg.GuestCreationEnabled = false

	before := countUsers(t, s)
	user, err := s.GuestSession(context.Background(), nil, guestTestSessionID(t))
	if err == nil {
		t.Fatal("expected an error with the kill switch off")
	}
	if user != nil {
		t.Fatalf("expected a nil user, got %+v", user)
	}
	after := countUsers(t, s)
	if after != before {
		t.Fatalf("expected zero users written, went from %d to %d", before, after)
	}
}

// TestGuestUsers_RateLimit proves T-02-01: a second GuestSession call from
// the same client key against a 1-per-minute/1-burst registry is refused
// and writes no row.
func TestGuestUsers_RateLimit(t *testing.T) {
	s := testAPI(t)
	s.limiter = ratelimit.NewRegistry(1, 1)
	sessionID := guestTestSessionID(t)

	before := countUsers(t, s)
	if _, err := s.GuestSession(context.Background(), nil, sessionID); err != nil {
		t.Fatalf("GuestSession() (first call) error = %v", err)
	}
	afterFirst := countUsers(t, s)
	if afterFirst != before+1 {
		t.Fatalf("expected exactly one row after the first call, got delta %d", afterFirst-before)
	}

	user, err := s.GuestSession(context.Background(), nil, sessionID)
	if err == nil {
		t.Fatal("expected the second call from the same client key to be rate limited")
	}
	if user != nil {
		t.Fatalf("expected a nil user on the rate-limited call, got %+v", user)
	}
	afterSecond := countUsers(t, s)
	if afterSecond != afterFirst {
		t.Fatalf("expected no additional row on the rate-limited call, went from %d to %d", afterFirst, afterSecond)
	}
}

// TestGuestUsers_EventDedup proves criterion 5: guest_session_created is
// idempotent under retry for the same (event, user, session) triple.
func TestGuestUsers_EventDedup(t *testing.T) {
	s := testAPI(t)
	sessionID := guestTestSessionID(t)
	t.Cleanup(func() { cleanupProductEvents(t, s.db, sessionID) })

	user, err := s.GuestSession(context.Background(), nil, sessionID)
	if err != nil {
		t.Fatalf("GuestSession() error = %v", err)
	}
	if got := countProductEvents(t, s.db, sessionID, "guest_session_created"); got != 1 {
		t.Fatalf("guest_session_created rows after creation = %d, want 1", got)
	}

	userID := user.ID
	s.recordProductEvent(context.Background(), ProductEvent{
		Name:      "guest_session_created",
		SessionID: sessionID,
		UserID:    &userID,
	}, false)

	if got := countProductEvents(t, s.db, sessionID, "guest_session_created"); got != 1 {
		t.Fatalf("guest_session_created rows after a repeated call = %d, want 1 (dedup failed)", got)
	}
}

func TestGuestUsers_DisplayNameValidation(t *testing.T) {
	originalHash := hashGuestSecret
	hashGuestSecret = func(secret string) (string, error) { return "fake-hash:" + secret, nil }
	t.Cleanup(func() { hashGuestSecret = originalHash })

	tests := []struct {
		name string
		raw  *string
		want *string
	}{
		{name: "null stays null"},
		{name: "whitespace becomes null", raw: stringPtr(" \t\n\r ")},
		{name: "truncates by rune", raw: stringPtr(strings.Repeat("界", 200)), want: stringPtr(strings.Repeat("界", 64))},
		{name: "strips controls", raw: stringPtr(" \x00A\nli\tce\u007f "), want: stringPtr("Alice")},
		{name: "keeps 64 multibyte runes", raw: stringPtr(strings.Repeat("界", 64)), want: stringPtr(strings.Repeat("界", 64))},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := testAPI(t)
			user, err := s.GuestSession(context.Background(), tt.raw, guestTestSessionID(t))
			if err != nil {
				t.Fatalf("GuestSession() error = %v", err)
			}
			if !reflect.DeepEqual(user.DisplayName, tt.want) {
				t.Fatalf("DisplayName = %#v, want %#v", user.DisplayName, tt.want)
			}

			var stored *string
			if err := s.db.QueryRow(`SELECT display_name FROM users WHERE uuid = $1`, user.ID).Scan(&stored); err != nil {
				t.Fatalf("scan display_name: %v", err)
			}
			if !reflect.DeepEqual(stored, tt.want) {
				t.Fatalf("stored display_name = %#v, want %#v", stored, tt.want)
			}
		})
	}

	t.Run("display names are deliberately non-unique", func(t *testing.T) {
		s := testAPI(t)
		name := "Dylan"
		first, err := s.GuestSession(context.Background(), &name, guestTestSessionID(t)+"-first")
		if err != nil {
			t.Fatalf("first GuestSession() error = %v", err)
		}
		second, err := s.GuestSession(context.Background(), &name, guestTestSessionID(t)+"-second")
		if err != nil {
			t.Fatalf("second GuestSession() error = %v", err)
		}
		if first.DisplayName == nil || second.DisplayName == nil || *first.DisplayName != name || *second.DisplayName != name {
			t.Fatalf("duplicate display names not preserved: first=%#v second=%#v", first.DisplayName, second.DisplayName)
		}
	})

	t.Run("normalization forms remain distinct and accepted", func(t *testing.T) {
		s := testAPI(t)
		composed := "Caf\u00e9"
		decomposed := "Cafe\u0301"
		first, err := s.GuestSession(context.Background(), &composed, guestTestSessionID(t)+"-composed")
		if err != nil {
			t.Fatalf("composed GuestSession() error = %v", err)
		}
		second, err := s.GuestSession(context.Background(), &decomposed, guestTestSessionID(t)+"-decomposed")
		if err != nil {
			t.Fatalf("decomposed GuestSession() error = %v", err)
		}
		if first.DisplayName == nil || second.DisplayName == nil || *first.DisplayName == *second.DisplayName {
			t.Fatalf("normalization forms were collapsed: first=%#v second=%#v", first.DisplayName, second.DisplayName)
		}
	})
}

func TestGuestUsers_ErrorCodes(t *testing.T) {
	originalHash := hashGuestSecret
	hashGuestSecret = func(secret string) (string, error) { return "fake-hash:" + secret, nil }
	t.Cleanup(func() { hashGuestSecret = originalHash })

	assertGuestSessionError := func(t *testing.T, err error) {
		t.Helper()
		if err == nil {
			t.Fatal("expected an error")
		}
		var gqlErr *gqlerror.Error
		if !errors.As(err, &gqlErr) {
			t.Fatalf("error type = %T, want *gqlerror.Error: %v", err, err)
		}
		if gqlErr.Extensions["code"] != string(ActivationCodeGuestSessionError) {
			t.Fatalf("extensions.code = %#v, want %q", gqlErr.Extensions["code"], ActivationCodeGuestSessionError)
		}
	}

	t.Run("empty session ID", func(t *testing.T) {
		s := testAPI(t)
		_, err := s.GuestSession(context.Background(), nil, " \t\n")
		assertGuestSessionError(t, err)
	})

	t.Run("kill switch", func(t *testing.T) {
		s := testAPI(t)
		s.cfg.GuestCreationEnabled = false
		_, err := s.GuestSession(context.Background(), nil, guestTestSessionID(t))
		assertGuestSessionError(t, err)
	})

	t.Run("rate limited", func(t *testing.T) {
		s := testAPI(t)
		s.limiter = ratelimit.NewRegistry(1, 1)
		sessionID := guestTestSessionID(t)
		if _, err := s.GuestSession(context.Background(), nil, sessionID); err != nil {
			t.Fatalf("first GuestSession() error = %v", err)
		}
		_, err := s.GuestSession(context.Background(), nil, sessionID)
		assertGuestSessionError(t, err)
	})

	t.Run("insert failure", func(t *testing.T) {
		s := testAPI(t)
		if err := s.db.Close(); err != nil {
			t.Fatalf("close test database: %v", err)
		}
		_, err := s.GuestSession(context.Background(), nil, guestTestSessionID(t))
		assertGuestSessionError(t, err)
		if strings.Contains(err.Error(), "database is closed") {
			t.Fatalf("client-facing error leaked database failure: %v", err)
		}
	})
}

func stringPtr(value string) *string {
	return &value
}

func TestGuestUsers_Refresh(t *testing.T) {
	useFastGuestHashes(t)

	t.Run("HappyPath", func(t *testing.T) {
		s := testAPI(t)
		guest, err := s.GuestSession(context.Background(), nil, guestTestSessionID(t))
		if err != nil {
			t.Fatalf("GuestSession() error = %v", err)
		}

		refreshed, err := s.RefreshGuestSession(context.Background(), *guest.GuestCredential)
		if err != nil {
			t.Fatalf("RefreshGuestSession() error = %v", err)
		}
		if refreshed == nil || refreshed.Token == nil || *refreshed.Token == "" {
			t.Fatalf("RefreshGuestSession() returned no token: %+v", refreshed)
		}
		authUser, err := parseAndValidateToken(*refreshed.Token)
		if err != nil {
			t.Fatalf("parseAndValidateToken() error = %v", err)
		}
		if authUser.ID != guest.ID || authUser.Username != guest.Username {
			t.Fatalf("refreshed claims = %+v, want ID=%q Username=%q", authUser, guest.ID, guest.Username)
		}
		if refreshed.GuestCredential == nil || *refreshed.GuestCredential != *guest.GuestCredential {
			t.Fatalf("GuestCredential rotated: got %#v, want %q", refreshed.GuestCredential, *guest.GuestCredential)
		}
		if refreshed.Password != nil {
			t.Fatal("RefreshGuestSession() returned Password material")
		}
	})

	t.Run("WrongSecret", func(t *testing.T) {
		s := testAPI(t)
		guest, err := s.GuestSession(context.Background(), nil, guestTestSessionID(t))
		if err != nil {
			t.Fatalf("GuestSession() error = %v", err)
		}
		parts := strings.SplitN(*guest.GuestCredential, ".", 2)
		_, wrongErr := s.RefreshGuestSession(context.Background(), parts[0]+".definitely-the-wrong-secret")
		assertActivationCode(t, wrongErr, ActivationCodeGuestSessionError)

		_, missingErr := s.RefreshGuestSession(context.Background(), uuid.NewString()+".definitely-the-wrong-secret")
		assertActivationCode(t, missingErr, ActivationCodeGuestSessionError)
		if wrongErr.Error() != missingErr.Error() {
			t.Fatalf("wrong-secret error = %q, missing-row error = %q; messages must be identical", wrongErr, missingErr)
		}
	})

	t.Run("MalformedCredential", func(t *testing.T) {
		for _, credential := range []string{"no-separator", uuid.NewString() + ".", ".secret", "not-a-uuid.secret"} {
			t.Run(credential, func(t *testing.T) {
				s := testAPI(t)
				if err := s.db.Close(); err != nil {
					t.Fatalf("close test database: %v", err)
				}
				_, err := s.RefreshGuestSession(context.Background(), credential)
				assertActivationCode(t, err, ActivationCodeGuestSessionError)
				if strings.Contains(err.Error(), "database is closed") {
					t.Fatalf("malformed credential reached the database: %v", err)
				}
			})
		}
	})

	t.Run("NonGuestRefused", func(t *testing.T) {
		s := testAPI(t)
		id := uuid.NewString()
		secret := "known-refresh-secret"
		hash, err := bcrypt.GenerateFromPassword([]byte(secret), bcrypt.MinCost)
		if err != nil {
			t.Fatalf("hash credential: %v", err)
		}
		if _, err := s.db.Exec(
			`INSERT INTO users (uuid, username, password, is_guest, guest_credential_hash) VALUES ($1, $2, $3, false, $4)`,
			id, uniqueUsername("refresh_non_guest"), "unused", string(hash),
		); err != nil {
			t.Fatalf("insert non-guest: %v", err)
		}
		_, err = s.RefreshGuestSession(context.Background(), id+"."+secret)
		assertActivationCode(t, err, ActivationCodeGuestSessionError)
	})
}

func TestGuestUsers_ExpiredRowRejected(t *testing.T) {
	useFastGuestHashes(t)
	s := testAPI(t)
	guest, err := s.GuestSession(context.Background(), nil, guestTestSessionID(t))
	if err != nil {
		t.Fatalf("GuestSession() error = %v", err)
	}

	comparisonNow := time.Now().UTC().Truncate(time.Microsecond).Add(321 * time.Microsecond)
	originalNow := guestSessionNow
	guestSessionNow = func() time.Time { return comparisonNow }
	t.Cleanup(func() { guestSessionNow = originalNow })

	setExpiry := func(value any) {
		t.Helper()
		if _, err := s.db.Exec(`UPDATE users SET expires_at = $1 WHERE uuid = $2`, value, guest.ID); err != nil {
			t.Fatalf("set expires_at = %#v: %v", value, err)
		}
	}

	setExpiry(comparisonNow.Add(-time.Second))
	_, err = s.RefreshGuestSession(context.Background(), *guest.GuestCredential)
	assertActivationCode(t, err, ActivationCodeGuestSessionError)

	setExpiry(comparisonNow.Add(time.Second))
	if _, err := s.RefreshGuestSession(context.Background(), *guest.GuestCredential); err != nil {
		t.Fatalf("future expires_at was refused: %v", err)
	}

	setExpiry(comparisonNow)
	_, err = s.RefreshGuestSession(context.Background(), *guest.GuestCredential)
	assertActivationCode(t, err, ActivationCodeGuestSessionError)

	setExpiry(nil)
	if _, err := s.RefreshGuestSession(context.Background(), *guest.GuestCredential); err != nil {
		t.Fatalf("NULL expires_at was refused: %v", err)
	}
}

func TestGuestUsers_ExpiredRowStillAuthorized(t *testing.T) {
	useFastGuestHashes(t)
	s := testAPI(t)
	guest, err := s.GuestSession(context.Background(), nil, guestTestSessionID(t))
	if err != nil {
		t.Fatalf("GuestSession() error = %v", err)
	}
	if _, err := s.db.Exec(`UPDATE users SET expires_at = $1 WHERE uuid = $2`, time.Now().Add(-time.Hour), guest.ID); err != nil {
		t.Fatalf("mark guest expired: %v", err)
	}

	authUser, err := parseAndValidateToken(*guest.Token)
	if err != nil {
		t.Fatalf("parseAndValidateToken() error = %v", err)
	}
	ctx := withAuth(context.Background(), authUser)
	gameID := "expired-row-live-token-" + guestTestSessionID(t)
	decklist := "1,Sol Ring"
	game, err := s.CreateGame(ctx, InputCreateGame{
		ID:   gameID,
		Turn: &InputTurn{Player: authUser.Username, Phase: "MAIN", Number: 1, Priority: authUser.Username},
		Players: []*InputBoardState{{
			UserID: authUser.ID, User: authUser.Username, GameID: gameID, Life: 40, Decklist: &decklist,
		}},
	})
	t.Cleanup(func() { _, _ = s.db.Exec(`DELETE FROM games WHERE id = $1`, gameID) })
	if err != nil {
		t.Fatalf("CreateGame() with a live JWT for a marked-expired row error = %v", err)
	}
	if game == nil || game.ID != gameID {
		t.Fatalf("CreateGame() = %+v, want game %q", game, gameID)
	}
}

func TestGuestUsers_TokenTTLBoundary(t *testing.T) {
	useFastGuestHashes(t)
	s := testAPI(t)
	guest, err := s.GuestSession(context.Background(), nil, guestTestSessionID(t))
	if err != nil {
		t.Fatalf("GuestSession() error = %v", err)
	}

	claims := &AuthClaims{}
	if _, _, err := jwt.NewParser().ParseUnverified(*guest.Token, claims); err != nil {
		t.Fatalf("ParseUnverified() error = %v", err)
	}
	if claims.IssuedAt == nil || claims.ExpiresAt == nil {
		t.Fatalf("token lacks issuance/expiry: %+v", claims)
	}
	if got := claims.ExpiresAt.Time.Sub(claims.IssuedAt.Time); got != 24*time.Hour {
		t.Fatalf("token TTL = %v, want 24h", got)
	}

	if _, err := parseTokenAt(*guest.Token, claims.IssuedAt.Time.Add(23*time.Hour+59*time.Minute)); err != nil {
		t.Fatalf("token rejected at 23h59m: %v", err)
	}
	if _, err := parseTokenAt(*guest.Token, claims.IssuedAt.Time.Add(24*time.Hour+time.Second)); err == nil {
		t.Fatal("token accepted at 24h00m01s")
	}
}

func useFastGuestHashes(t *testing.T) {
	t.Helper()
	originalHash := hashGuestSecret
	hashGuestSecret = func(secret string) (string, error) {
		hash, err := bcrypt.GenerateFromPassword([]byte(secret), bcrypt.MinCost)
		return string(hash), err
	}
	t.Cleanup(func() { hashGuestSecret = originalHash })
}

func parseTokenAt(tokenString string, at time.Time) (*AuthClaims, error) {
	secret, err := jwtSecret()
	if err != nil {
		return nil, err
	}
	claims := &AuthClaims{}
	parsed, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		return secret, nil
	}, jwt.WithTimeFunc(func() time.Time { return at }))
	if err != nil {
		return nil, err
	}
	if !parsed.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

func TestGuestUsers_Claim(t *testing.T) {
	useFastGuestHashes(t)
	useFastClaimHashes(t)

	t.Run("HappyPath", func(t *testing.T) {
		s := testAPI(t)
		guest, err := s.GuestSession(context.Background(), stringPtr("Table Mage"), guestTestSessionID(t))
		if err != nil {
			t.Fatalf("GuestSession() error = %v", err)
		}
		beforeUsers := countUsers(t, s)
		gameID := createGameForGuest(t, s, guest)

		claimSessionID := guestTestSessionID(t)
		t.Cleanup(func() { cleanupProductEvents(t, s.db, claimSessionID) })
		claimedUsername := uniqueUsername("claimed")
		claimedPassword := "permanent-password"
		claimed, err := s.ClaimGuestAccount(guestAuthContext(t, guest), claimedUsername, claimedPassword, claimSessionID)
		if err != nil {
			t.Fatalf("ClaimGuestAccount() error = %v", err)
		}
		if countUsers(t, s) != beforeUsers {
			t.Fatalf("claim changed users row count: before=%d after=%d", beforeUsers, countUsers(t, s))
		}
		if claimed == nil || claimed.ID != guest.ID || claimed.Username != claimedUsername {
			t.Fatalf("claimed user = %+v, want same ID %q and username %q", claimed, guest.ID, claimedUsername)
		}
		if claimed.IsGuest == nil || *claimed.IsGuest || claimed.GuestCredential != nil || claimed.Password != nil {
			t.Fatalf("claimed response retained guest or secret state: %+v", claimed)
		}
		if claimed.Token == nil || *claimed.Token == "" {
			t.Fatal("claim returned no full-session token")
		}

		row := readGuestClaimRow(t, s, guest.ID)
		if row.Username != claimedUsername || row.IsGuest || row.ExpiresAt.Valid || row.CredentialHash.Valid {
			t.Fatalf("claimed row has wrong identity state: %+v", row)
		}
		if !checkPasswordHash(claimedPassword, row.Password) {
			t.Fatal("claimed password does not validate through checkPasswordHash")
		}

		claimedAuth, err := parseAndValidateToken(*claimed.Token)
		if err != nil {
			t.Fatalf("parse claimed token: %v", err)
		}
		game, err := s.GetGame(withAuth(context.Background(), claimedAuth), gameID)
		if err != nil {
			t.Fatalf("GetGame() after claim error = %v", err)
		}
		if game == nil || game.ID != gameID {
			t.Fatalf("GetGame() after claim = %+v, want %q", game, gameID)
		}
		if got := countProductEvents(t, s.db, claimSessionID, "account_claimed"); got != 1 {
			t.Fatalf("account_claimed rows = %d, want 1 (a stored authoritative event proves clientSubmitted=false)", got)
		}
	})

	t.Run("UsernameTaken", func(t *testing.T) {
		s := testAPI(t)
		taken := uniqueUsername("claim_taken")
		if _, err := s.db.Exec(`INSERT INTO users (uuid, username, password, is_guest) VALUES ($1, $2, $3, false)`, uuid.NewString(), taken, "unused"); err != nil {
			t.Fatalf("seed taken username: %v", err)
		}
		guest, err := s.GuestSession(context.Background(), nil, guestTestSessionID(t))
		if err != nil {
			t.Fatalf("GuestSession() error = %v", err)
		}
		before := readGuestClaimRow(t, s, guest.ID)

		claimed, err := s.ClaimGuestAccount(guestAuthContext(t, guest), taken, "new-password", guestTestSessionID(t))
		if err == nil || activationProductMessage(t, err) != "That username is already taken. Try another one." {
			t.Fatalf("ClaimGuestAccount() error = %v, want Signup's taken-name message", err)
		}
		if claimed != nil {
			t.Fatalf("ClaimGuestAccount() returned user on conflict: %+v", claimed)
		}
		if after := readGuestClaimRow(t, s, guest.ID); !reflect.DeepEqual(after, before) {
			t.Fatalf("guest row changed on username conflict:\nbefore=%+v\nafter=%+v", before, after)
		}
	})

	t.Run("NotAGuest", func(t *testing.T) {
		s := testAPI(t)
		id := uuid.NewString()
		username := uniqueUsername("claim_real")
		if _, err := s.db.Exec(`INSERT INTO users (uuid, username, password, is_guest) VALUES ($1, $2, $3, false)`, id, username, "unchanged-password"); err != nil {
			t.Fatalf("insert non-guest: %v", err)
		}
		before := readGuestClaimRow(t, s, id)
		claimed, err := s.ClaimGuestAccount(authCtxWithID(id, username), uniqueUsername("should_not_apply"), "new-password", guestTestSessionID(t))
		assertActivationCode(t, err, ActivationCodeGuestSessionError)
		if claimed != nil {
			t.Fatalf("ClaimGuestAccount() returned non-guest: %+v", claimed)
		}
		if after := readGuestClaimRow(t, s, id); !reflect.DeepEqual(after, before) {
			t.Fatalf("non-guest row changed:\nbefore=%+v\nafter=%+v", before, after)
		}
	})

	t.Run("EmptyPassword", func(t *testing.T) {
		s := testAPI(t)
		guest, err := s.GuestSession(context.Background(), nil, guestTestSessionID(t))
		if err != nil {
			t.Fatalf("GuestSession() error = %v", err)
		}
		before := readGuestClaimRow(t, s, guest.ID)
		claimed, err := s.ClaimGuestAccount(guestAuthContext(t, guest), uniqueUsername("claim_empty_password"), "", guestTestSessionID(t))
		if err == nil || activationProductMessage(t, err) != "must provide a password" {
			t.Fatalf("ClaimGuestAccount() error = %v, want Signup's empty-password message", err)
		}
		if claimed != nil {
			t.Fatalf("ClaimGuestAccount() returned user for empty password: %+v", claimed)
		}
		if after := readGuestClaimRow(t, s, guest.ID); !reflect.DeepEqual(after, before) {
			t.Fatalf("guest row changed for empty password:\nbefore=%+v\nafter=%+v", before, after)
		}
	})

	t.Run("EventOnce", func(t *testing.T) {
		s := testAPI(t)
		guest, err := s.GuestSession(context.Background(), nil, guestTestSessionID(t))
		if err != nil {
			t.Fatalf("GuestSession() error = %v", err)
		}
		sessionID := guestTestSessionID(t)
		t.Cleanup(func() { cleanupProductEvents(t, s.db, sessionID) })
		claimed, err := s.ClaimGuestAccount(guestAuthContext(t, guest), uniqueUsername("claim_event"), "new-password", sessionID)
		if err != nil {
			t.Fatalf("ClaimGuestAccount() error = %v", err)
		}
		if got := countProductEvents(t, s.db, sessionID, "account_claimed"); got != 1 {
			t.Fatalf("account_claimed rows after claim = %d, want 1", got)
		}
		userID := claimed.ID
		s.recordProductEvent(context.Background(), ProductEvent{Name: "account_claimed", SessionID: sessionID, UserID: &userID}, false)
		if got := countProductEvents(t, s.db, sessionID, "account_claimed"); got != 1 {
			t.Fatalf("account_claimed rows after retry = %d, want 1", got)
		}
	})
}

func activationProductMessage(t *testing.T, err error) string {
	t.Helper()
	var gqlErr *gqlerror.Error
	if !errors.As(err, &gqlErr) {
		t.Fatalf("error type = %T, want *gqlerror.Error: %v", err, err)
	}
	return gqlErr.Message
}

type guestClaimRow struct {
	Username       string
	Password       string
	IsGuest        bool
	ExpiresAt      sql.NullTime
	DisplayName    sql.NullString
	CredentialHash sql.NullString
}

func readGuestClaimRow(t *testing.T, s *graphQLServer, id string) guestClaimRow {
	t.Helper()
	var row guestClaimRow
	if err := s.db.QueryRow(`
		SELECT username, password, is_guest, expires_at, display_name, guest_credential_hash
		FROM users WHERE uuid = $1
	`, id).Scan(&row.Username, &row.Password, &row.IsGuest, &row.ExpiresAt, &row.DisplayName, &row.CredentialHash); err != nil {
		t.Fatalf("read user row %q: %v", id, err)
	}
	return row
}

func guestAuthContext(t *testing.T, guest *User) context.Context {
	t.Helper()
	if guest == nil || guest.Token == nil {
		t.Fatal("guest has no token")
	}
	authUser, err := parseAndValidateToken(*guest.Token)
	if err != nil {
		t.Fatalf("parse guest token: %v", err)
	}
	return withAuth(context.Background(), authUser)
}

func createGameForGuest(t *testing.T, s *graphQLServer, guest *User) string {
	t.Helper()
	ctx := guestAuthContext(t, guest)
	gameID := "claim-preserves-game-" + guestTestSessionID(t)
	decklist := "1,Sol Ring"
	game, err := s.CreateGame(ctx, InputCreateGame{
		ID:   gameID,
		Turn: &InputTurn{Player: guest.Username, Phase: "MAIN", Number: 1, Priority: guest.Username},
		Players: []*InputBoardState{{
			UserID: guest.ID, User: guest.Username, GameID: gameID, Life: 40, Decklist: &decklist,
		}},
	})
	if err != nil {
		t.Fatalf("CreateGame() error = %v", err)
	}
	if game == nil || game.ID != gameID {
		t.Fatalf("CreateGame() = %+v, want %q", game, gameID)
	}
	t.Cleanup(func() { _, _ = s.db.Exec(`DELETE FROM games WHERE id = $1`, gameID) })
	return gameID
}

func useFastClaimHashes(t *testing.T) {
	t.Helper()
	originalHash := hashClaimPassword
	hashClaimPassword = func(password string) (string, error) {
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
		return string(hash), err
	}
	t.Cleanup(func() { hashClaimPassword = originalHash })
}

// TestMigrations_GuestUsers proves the migration itself: the four new
// columns exist with the right shape, is_guest defaults false, and the
// production/test migration pair is byte-identical.
func TestMigrations_GuestUsers(t *testing.T) {
	s := testAPI(t)

	type columnSpec struct {
		name     string
		dataType string
	}
	want := []columnSpec{
		{"is_guest", "boolean"},
		{"expires_at", "timestamp with time zone"},
		{"display_name", "character varying"},
		{"guest_credential_hash", "character varying"},
	}
	for _, col := range want {
		var dataType string
		err := s.db.QueryRow(
			`SELECT data_type FROM information_schema.columns WHERE table_name = 'users' AND column_name = $1`,
			col.name,
		).Scan(&dataType)
		if err != nil {
			t.Fatalf("column %q: %v", col.name, err)
		}
		if dataType != col.dataType {
			t.Errorf("column %q: data_type = %q, want %q", col.name, dataType, col.dataType)
		}
	}

	// A row inserted without specifying is_guest must default to false.
	defaultUsername := uniqueUsername("guestdefaultcheck")
	if _, err := s.db.Exec(
		`INSERT INTO users (uuid, username, password) VALUES ($1, $2, $3)`,
		defaultUsername, defaultUsername, "irrelevant-hash",
	); err != nil {
		t.Fatalf("insert row without is_guest: %v", err)
	}
	var isGuestDefault bool
	if err := s.db.QueryRow(`SELECT is_guest FROM users WHERE uuid = $1`, defaultUsername).Scan(&isGuestDefault); err != nil {
		t.Fatalf("scan default is_guest: %v", err)
	}
	if isGuestDefault {
		t.Fatal("expected is_guest to default to false")
	}

	prodBytes, err := os.ReadFile("../persistence/migrations/20260809120000_guest_users.up.sql")
	if err != nil {
		t.Fatalf("read production migration: %v", err)
	}
	testBytes, err := os.ReadFile("../persistence/migrations_test/20260809120000_guest_users.up.sql")
	if err != nil {
		t.Fatalf("read test migration: %v", err)
	}
	if string(prodBytes) != string(testBytes) {
		t.Fatal("expected the production and test migration files to be byte-identical")
	}
}

func isGuestAdjective(word string) bool {
	for _, a := range guestAdjectives {
		if a == word {
			return true
		}
	}
	return false
}

func isGuestNoun(word string) bool {
	for _, n := range guestNouns {
		if n == word {
			return true
		}
	}
	return false
}
