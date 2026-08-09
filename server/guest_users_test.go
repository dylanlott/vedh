package server

import (
	"context"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/openmtg/edh-go/pkg/ratelimit"
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
