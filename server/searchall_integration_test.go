package server

import (
	"bufio"
	"context"
	"os"
	"strings"
	"testing"
)

// Test_graphQLServer_SearchAll_JaradDeck tests the SearchAll function against a
// real-world decklist (Jarad, Golgari Lich Lord). It allows a small fraction of
// misses since the cards table may not have all printings or may use
// different naming for some cards.
func Test_graphQLServer_SearchAll_JaradDeck(t *testing.T) {
	f, err := os.Open("../test/decklists/jarad.csv")
	if err != nil {
		t.Fatalf("failed to open decklist: %v", err)
	}
	defer f.Close()

	s := testAPI(t)
	scanner := bufio.NewScanner(f)
	var rawNames []string
	seen := map[string]bool{}
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ",", 2)
		if len(parts) < 2 {
			continue
		}
		name := strings.TrimSpace(parts[1])
		name = strings.Trim(name, "\"")
		if seen[name] {
			continue
		}
		seen[name] = true
		rawNames = append(rawNames, name)
	}

	var failures []string
	for _, name := range rawNames {
		got, err := s.SearchAll(context.Background(), &name, nil, nil, nil)
		if err != nil {
			t.Logf("error searching %q: %v", name, err)
			failures = append(failures, name)
			continue
		}
		if len(got) == 0 {
			t.Logf("no results for %q", name)
			failures = append(failures, name)
		} else {
			t.Logf("found card: %+v\n", got)
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scanner error: %v", err)
	}

	// Allow a small fraction of misses (10%) since `cards` may omit some
	// tokens/printings or use different naming for some entries.
	allowedMissing := int(float64(len(rawNames)) * 0.10)
	if allowedMissing < 1 {
		allowedMissing = 1
	}
	if len(failures) > allowedMissing {
		end := 5
		if len(failures) < end {
			end = len(failures)
		}
		t.Fatalf("SearchAll failed for %d names (allowed %d), examples: %v", len(failures), allowedMissing, failures[:end])
	}
}
