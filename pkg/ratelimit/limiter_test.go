package ratelimit

import (
	"testing"
	"time"
)

// TestRegistry_BurstThenLimited proves the base token-bucket behavior: a
// burst of requests up to the burst size is allowed, and the next one is
// limited.
func TestRegistry_BurstThenLimited(t *testing.T) {
	r := NewRegistry(60, 3)

	for i := 0; i < 3; i++ {
		if !r.Allow(SurfaceDeckImport, "client-a") {
			t.Fatalf("request %d: expected allowed within burst", i)
		}
	}
	if r.Allow(SurfaceDeckImport, "client-a") {
		t.Fatal("expected the request beyond the burst to be limited")
	}
}

// TestRegistry_PerKeyIsolation proves two distinct client keys never share
// a bucket: exhausting one key's burst must not affect another key on the
// same surface.
func TestRegistry_PerKeyIsolation(t *testing.T) {
	r := NewRegistry(60, 1)

	if !r.Allow(SurfaceDeckImport, "client-a") {
		t.Fatal("expected client-a's first request to be allowed")
	}
	if r.Allow(SurfaceDeckImport, "client-a") {
		t.Fatal("expected client-a's second request to be limited")
	}
	if !r.Allow(SurfaceDeckImport, "client-b") {
		t.Fatal("expected client-b's first request to be allowed despite client-a's exhausted bucket")
	}
}

// TestRegistry_PerSurfaceIsolation proves two distinct surfaces never share
// a bucket: exhausting one surface's burst for a client must not affect the
// same client on a different surface.
func TestRegistry_PerSurfaceIsolation(t *testing.T) {
	r := NewRegistry(60, 1)

	if !r.Allow(SurfaceDeckImport, "client-a") {
		t.Fatal("expected the first deck-import request to be allowed")
	}
	if r.Allow(SurfaceDeckImport, "client-a") {
		t.Fatal("expected the second deck-import request to be limited")
	}
	if !r.Allow(SurfaceProductEvent, "client-a") {
		t.Fatal("expected the same client's product-event request to be allowed on a different surface")
	}
}

// TestRegistry_IdleEvictionGrantsFreshBucket proves that an entry idle past
// idleWindow is evicted, by advancing the injected clock rather than
// sleeping, and that a subsequent request from the same key is allowed
// with a fresh bucket rather than remaining limited.
func TestRegistry_IdleEvictionGrantsFreshBucket(t *testing.T) {
	r := NewRegistry(60, 1)
	current := time.Now()
	r.now = func() time.Time { return current }

	if !r.Allow(SurfaceDeckImport, "client-a") {
		t.Fatal("expected the first request to be allowed")
	}
	if r.Allow(SurfaceDeckImport, "client-a") {
		t.Fatal("expected the second request to be limited")
	}

	current = current.Add(idleWindow + time.Second)
	if !r.Allow(SurfaceDeckImport, "client-a") {
		t.Fatal("expected a request after the idle window to be allowed with a fresh bucket")
	}
}

// TestRegistry_SweepReturnsToZero proves the actual bounded-growth claim:
// after every entry has gone idle past idleWindow, a sweep leaves the
// registry with zero entries.
func TestRegistry_SweepReturnsToZero(t *testing.T) {
	r := NewRegistry(60, 1)
	current := time.Now()
	r.now = func() time.Time { return current }

	r.Allow(SurfaceDeckImport, "client-a")
	r.Allow(SurfaceDeckImport, "client-b")
	r.Allow(SurfaceProductEvent, "client-a")

	r.mu.Lock()
	if got := len(r.entries); got != 3 {
		r.mu.Unlock()
		t.Fatalf("expected 3 entries before the sweep, got %d", got)
	}
	r.mu.Unlock()

	current = current.Add(idleWindow + time.Second)

	r.mu.Lock()
	r.sweepLocked(current)
	got := len(r.entries)
	r.mu.Unlock()

	if got != 0 {
		t.Fatalf("expected 0 entries after a full sweep with every entry idle, got %d", got)
	}
}
