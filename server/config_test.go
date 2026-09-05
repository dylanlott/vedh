package server

import "testing"

func validTestConf() Conf {
	return Conf{
		PostgresURL:               "postgres://user:pass@localhost:5432/db?sslmode=disable",
		DefaultPort:               8080,
		JWTSecret:                 "01234567890123456789012345678901",
		AllowedOrigins:            "http://localhost:5173",
		DeckImportRatePerMinute:   30,
		DeckImportRateBurst:       10,
		GuestSessionRatePerMinute: 20,
		GuestSessionRateBurst:     5,
	}
}

func TestConfValidateAcceptsValidConfiguration(t *testing.T) {
	if err := validTestConf().Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestConfValidateRejectsWeakJWTSecret(t *testing.T) {
	cfg := validTestConf()
	cfg.JWTSecret = "too-short"
	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate() unexpectedly accepted a weak JWT_SECRET")
	}
}

func TestConfValidateRequiresMetricsTokenWhenEnabled(t *testing.T) {
	cfg := validTestConf()
	cfg.MetricsEnabled = true
	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate() unexpectedly accepted metrics without METRICS_TOKEN")
	}
}

func TestConfValidateRequiresProviderAllowlistWhenEnabled(t *testing.T) {
	cfg := validTestConf()
	cfg.DeckProviderEnabled = true
	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate() unexpectedly accepted provider fetches without an allowlist")
	}
}
