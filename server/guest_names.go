package server

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// guestAdjectives and guestNouns are the curated MTG-flavored word lists
// D-2.5 requires. Every entry has been reviewed to be Commander-flavored
// and safe to show a stranger: no slurs, no insults, nothing absurd. The
// full cross product (40+ * 40+ = 1600+ pairings) is the fixed namespace
// D-2.6/D-2.1 require -- because guest rows are immortal (D-2.1) names are
// never recycled, so generateGuestName's numeric-suffix overflow strategy
// below is the expected long-run path once the namespace nears exhaustion,
// not an edge case.
var guestAdjectives = []string{
	"Brave", "Mighty", "Swift", "Ancient", "Radiant", "Cunning", "Valiant",
	"Fierce", "Noble", "Wandering", "Eternal", "Vigilant", "Resolute",
	"Gilded", "Verdant", "Storming", "Silent", "Blazing", "Frozen",
	"Towering", "Wily", "Steadfast", "Luminous", "Wild", "Sturdy",
	"Daring", "Serene", "Thundering", "Shimmering", "Stalwart", "Roaring",
	"Nimble", "Untamed", "Sunlit", "Moonlit", "Boundless", "Unbroken",
	"Gallant", "Restless", "Tireless", "Emerald", "Crimson", "Azure",
	"Golden", "Obsidian", "Feral", "Watchful", "Spirited", "Undaunted",
	"Relentless",
}

var guestNouns = []string{
	"Sliver", "Wurm", "Griffin", "Phoenix", "Golem", "Hydra", "Dragon",
	"Sphinx", "Wyvern", "Elemental", "Angel", "Djinn", "Basilisk",
	"Chimera", "Kraken", "Manticore", "Treefolk", "Sentinel", "Warden",
	"Ranger", "Artificer", "Druid", "Shaman", "Wizard", "Knight",
	"Paladin", "Berserker", "Cleric", "Scout", "Rogue", "Elf", "Giant",
	"Spirit", "Construct", "Beast", "Serpent", "Falcon", "Bear", "Wolf",
	"Hawk", "Panther", "Cat", "Turtle", "Fox", "Otter", "Badger",
	"Lynx", "Owl", "Raven", "Elephant",
}

// guestNameOverflowMaxAttempts bounds generateGuestName's numeric-suffix
// overflow loop, so a pathological caller can never spin forever even
// though the loop is mathematically unbounded in principle.
const guestNameOverflowMaxAttempts = 1000

// drawGuestNamePair is the injectable draw seam D-2.5's collision test
// forces: a package-level function variable, defaulting to a real
// crypto/rand draw, that TestGuestUsers_Create/CollisionRetries reassigns
// to force a specific first pairing. Production code must never assign to
// this variable outside of a test.
var drawGuestNamePair = cryptoDrawGuestNamePair

// cryptoDrawGuestNamePair draws one adjective and one noun using
// crypto/rand -- never math/rand, which server/backronym.go uses for
// unrelated, non-security-relevant flavor text and is explicitly not a
// pattern to follow here (guest usernames back a real, indefinitely-lived
// authentication identity, per D-2.1).
func cryptoDrawGuestNamePair() (string, string, error) {
	adjIdx, err := rand.Int(rand.Reader, big.NewInt(int64(len(guestAdjectives))))
	if err != nil {
		return "", "", fmt.Errorf("guest name: draw adjective: %w", err)
	}
	nounIdx, err := rand.Int(rand.Reader, big.NewInt(int64(len(guestNouns))))
	if err != nil {
		return "", "", fmt.Errorf("guest name: draw noun: %w", err)
	}
	return guestAdjectives[adjIdx.Int64()], guestNouns[nounIdx.Int64()], nil
}

// generateGuestName draws an adjective-noun pairing via drawGuestNamePair,
// joined with a single space. attempt is 0 for the first draw a caller
// makes and increments on each subsequent retry after a
// username_unique collision (server/guest_users.go); once attempt exceeds
// the size of the curated cross product, a numeric suffix is appended
// (D-2.5's overflow strategy for the fixed, never-recycled namespace,
// D-2.6). attempt is bounded by guestNameOverflowMaxAttempts so a caller
// cannot be made to loop forever.
func generateGuestName(attempt int) (string, error) {
	if attempt < 0 || attempt > guestNameOverflowMaxAttempts {
		return "", fmt.Errorf("guest name: attempt %d out of bounds", attempt)
	}
	adjective, noun, err := drawGuestNamePair()
	if err != nil {
		return "", err
	}
	crossProductSize := len(guestAdjectives) * len(guestNouns)
	if attempt < crossProductSize {
		return adjective + " " + noun, nil
	}
	// Overflow: the curated cross product has been exhausted by however
	// many prior attempts already failed on a collision. Append a numeric
	// suffix derived from the attempt number so retries are deterministic
	// in shape (still combined with a freshly drawn pairing, so the
	// number alone is never the only source of entropy).
	suffix := attempt - crossProductSize + 1
	return fmt.Sprintf("%s %s %d", adjective, noun, suffix), nil
}
