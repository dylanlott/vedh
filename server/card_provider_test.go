package server

import "testing"

func TestCardProviderRegistryRejectsUnknownGameType(t *testing.T) {
	registry := NewCardProviderRegistry()

	_, err := registry.ProviderForGameType("pokemon")

	if err == nil {
		t.Fatal("expected missing provider error")
	}
}
