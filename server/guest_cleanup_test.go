package server

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestGuestUsers_CleanupPreservesActiveGames(t *testing.T) {
	useFastGuestHashes(t)
	s := testAPI(t)
	past := time.Now().Add(-time.Hour)

	protected := createCleanupGuest(t, s, "protected")
	deletable := createCleanupGuest(t, s, "deletable")
	durable := createCleanupGuest(t, s, "durable")
	if _, err := s.db.Exec(`UPDATE users SET expires_at = $1 WHERE uuid IN ($2, $3)`, past, protected.ID, deletable.ID); err != nil {
		t.Fatalf("mark cleanup guests: %v", err)
	}

	gameID := "cleanup-protected-" + guestTestSessionID(t)
	payload := `{"guest_uuid":"` + protected.ID + `","guest_username":"` + protected.Username + `"}`
	if _, err := s.db.Exec(`INSERT INTO games (id, payload) VALUES ($1, $2::jsonb)`, gameID, payload); err != nil {
		t.Fatalf("insert protecting game payload: %v", err)
	}
	t.Cleanup(func() { _, _ = s.db.Exec(`DELETE FROM games WHERE id = $1`, gameID) })

	deleted, err := s.CleanupExpiredGuests(context.Background())
	if err != nil {
		t.Fatalf("CleanupExpiredGuests() error = %v", err)
	}
	if deleted != 1 {
		t.Fatalf("CleanupExpiredGuests() deleted = %d, want 1", deleted)
	}
	if !userRowExists(t, s, protected.ID) {
		t.Fatal("cleanup deleted a guest referenced by a game payload")
	}
	if userRowExists(t, s, deletable.ID) {
		t.Fatal("cleanup preserved an expired, unreferenced guest")
	}
	if !userRowExists(t, s, durable.ID) {
		t.Fatal("cleanup deleted a guest with NULL expires_at")
	}
}

func TestGuestUsers_CleanupIdempotent(t *testing.T) {
	useFastGuestHashes(t)
	s := testAPI(t)
	guest := createCleanupGuest(t, s, "idempotent")
	if _, err := s.db.Exec(`UPDATE users SET expires_at = $1 WHERE uuid = $2`, time.Now().Add(-time.Minute), guest.ID); err != nil {
		t.Fatalf("mark cleanup guest: %v", err)
	}

	first, err := s.CleanupExpiredGuests(context.Background())
	if err != nil {
		t.Fatalf("first CleanupExpiredGuests() error = %v", err)
	}
	second, err := s.CleanupExpiredGuests(context.Background())
	if err != nil {
		t.Fatalf("second CleanupExpiredGuests() error = %v", err)
	}
	if first != 1 || second != 0 {
		t.Fatalf("cleanup counts = (%d, %d), want (1, 0)", first, second)
	}
}

func TestGuestUsers_CleanupNeverTouchesRealUsers(t *testing.T) {
	s := testAPI(t)
	id := uuid.NewString()
	if _, err := s.db.Exec(
		`INSERT INTO users (uuid, username, password, is_guest, expires_at) VALUES ($1, $2, $3, false, $4)`,
		id, uniqueUsername("cleanup_real"), "unused", time.Now().Add(-time.Hour),
	); err != nil {
		t.Fatalf("insert real user: %v", err)
	}
	deleted, err := s.CleanupExpiredGuests(context.Background())
	if err != nil {
		t.Fatalf("CleanupExpiredGuests() error = %v", err)
	}
	if deleted != 0 || !userRowExists(t, s, id) {
		t.Fatalf("cleanup touched a real user: deleted=%d exists=%t", deleted, userRowExists(t, s, id))
	}
}

func TestGuestUsers_CleanupNoScheduledWork(t *testing.T) {
	useFastGuestHashes(t)
	s := testAPI(t)
	guest := createCleanupGuest(t, s, "durable_shape")
	deleted, err := s.CleanupExpiredGuests(context.Background())
	if err != nil {
		t.Fatalf("CleanupExpiredGuests() error = %v", err)
	}
	if deleted != 0 || !userRowExists(t, s, guest.ID) {
		t.Fatalf("production-shape guest was cleanup work: deleted=%d exists=%t", deleted, userRowExists(t, s, guest.ID))
	}
}

func createCleanupGuest(t *testing.T, s *graphQLServer, suffix string) *User {
	t.Helper()
	guest, err := s.GuestSession(context.Background(), nil, guestTestSessionID(t)+"-"+suffix)
	if err != nil {
		t.Fatalf("GuestSession(%q) error = %v", suffix, err)
	}
	return guest
}

func userRowExists(t *testing.T, s *graphQLServer, id string) bool {
	t.Helper()
	var exists bool
	if err := s.db.QueryRow(`SELECT EXISTS (SELECT 1 FROM users WHERE uuid = $1)`, id).Scan(&exists); err != nil {
		t.Fatalf("check user %q existence: %v", id, err)
	}
	return exists
}
