package server

import (
	"errors"
	"math/rand"
	"strings"
	"time"
)

// DefaultDictionary is a baseline word set you can extend/override.
// Keys are uppercase ASCII letters.
var DefaultDictionary = map[rune][]string{
	'V': {"Visual", "Virtual", "Volatile", "Verified", "Versatile", "Versioned", "Vivid", "Vertical"},
	'E': {"Encrypted", "Experimental", "Enhanced", "Eternal", "Elastic", "Esoteric", "Extensible", "Evolved"},
	'D': {"Deck", "Distributed", "Deterministic", "Dynamic", "Decentralized", "Diverse", "Digital", "Dedicated"},
	'H': {"Hub", "Handler", "Hypermedia", "Hypervisor", "Heighliner", "Hologram", "Host", "Hybrid"},
}

// GenerateVEDH returns every valid backronym for "VEDH" using dict.
// Example output: "Virtual Encrypted Deck Hypervisor".
func GenerateVEDH(dict map[rune][]string) ([]string, error) {
	return GenerateForAcronym(dict, "VEDH")
}

// GenerateForAcronym generates every combination for an acronym (e.g. "VEDH").
// Dict keys must match the acronym letters (case-insensitive).
func GenerateForAcronym(dict map[rune][]string, acronym string) ([]string, error) {
	if strings.TrimSpace(acronym) == "" {
		return nil, errors.New("acronym must not be empty")
	}

	letters := []rune(strings.ToUpper(acronym))
	for _, ch := range letters {
		if _, ok := dict[ch]; !ok || len(dict[ch]) == 0 {
			return nil, errors.New("dictionary missing words for letter: " + string(ch))
		}
	}

	results := make([]string, 0, estimateCount(dict, letters))

	var dfs func(pos int, current []string)
	dfs = func(pos int, current []string) {
		if pos == len(letters) {
			results = append(results, strings.Join(current, " "))
			return
		}
		ch := letters[pos]
		for _, w := range dict[ch] {
			dfs(pos+1, append(current, w))
		}
	}

	dfs(0, nil)
	return results, nil
}

// SampleVEDH returns up to n randomly sampled VEDH backronyms without enumerating all.
// This is useful if the cartesian product is huge.
func SampleVEDH(dict map[rune][]string, n int) ([]string, error) {
	return SampleForAcronym(dict, "VEDH", n, rand.New(rand.NewSource(time.Now().UnixNano())))
}

// SampleForAcronym returns up to n sampled backronyms for the given acronym using rng.
// Sampling is with replacement (duplicates possible) unless you dedupe yourself.
// If you want deterministic sampling, pass a seeded rng.
func SampleForAcronym(dict map[rune][]string, acronym string, n int, rng *rand.Rand) ([]string, error) {
	if n < 0 {
		return nil, errors.New("n must be >= 0")
	}
	if rng == nil {
		return nil, errors.New("rng must not be nil")
	}
	if strings.TrimSpace(acronym) == "" {
		return nil, errors.New("acronym must not be empty")
	}

	letters := []rune(strings.ToUpper(acronym))
	for _, ch := range letters {
		ws, ok := dict[ch]
		if !ok || len(ws) == 0 {
			return nil, errors.New("dictionary missing words for letter: " + string(ch))
		}
	}

	out := make([]string, 0, n)
	for i := 0; i < n; i++ {
		parts := make([]string, 0, len(letters))
		for _, ch := range letters {
			ws := dict[ch]
			parts = append(parts, ws[rng.Intn(len(ws))])
		}
		out = append(out, strings.Join(parts, " "))
	}

	return out, nil
}

// estimateCount pre-allocates for GenerateForAcronym.
func estimateCount(dict map[rune][]string, letters []rune) int {
	// Multiply lengths; cap to avoid int overflow nastiness.
	const capCount = 2_000_000
	count := 1
	for _, ch := range letters {
		l := len(dict[ch])
		if l == 0 {
			return 0
		}
		if count > capCount/l {
			return capCount
		}
		count *= l
	}
	return count
}
