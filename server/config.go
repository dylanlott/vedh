package server

import (
	"fmt"
	"net/url"
	"strings"
)

// Validate checks semantic configuration constraints that envconfig cannot
// express. It is called during production startup so a bad deployment fails
// before opening a listener.
func (c Conf) Validate() error {
	if strings.TrimSpace(c.PostgresURL) == "" {
		return fmt.Errorf("DATABASE_URL must not be empty")
	}
	dsn, err := url.Parse(c.PostgresURL)
	if err != nil || dsn.Host == "" || (dsn.Scheme != "postgres" && dsn.Scheme != "postgresql") {
		return fmt.Errorf("DATABASE_URL must be a postgres:// URL with a host")
	}
	if c.DefaultPort < 1 || c.DefaultPort > 65535 {
		return fmt.Errorf("PORT must be between 1 and 65535")
	}
	if len(strings.TrimSpace(c.JWTSecret)) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters")
	}
	if len(parseAllowedOrigins(c.AllowedOrigins)) == 0 {
		return fmt.Errorf("ALLOWED_ORIGINS must contain at least one valid http or https origin")
	}
	if c.MetricsEnabled && strings.TrimSpace(c.MetricsToken) == "" {
		return fmt.Errorf("METRICS_TOKEN is required when METRICS_ENABLED=true")
	}
	if c.DeckImportRatePerMinute < 1 || c.DeckImportRateBurst < 1 {
		return fmt.Errorf("deck import rate limits must be positive")
	}
	if c.GuestSessionRatePerMinute < 1 || c.GuestSessionRateBurst < 1 {
		return fmt.Errorf("guest session rate limits must be positive")
	}
	if c.DeckProviderEnabled && len(parseAllowedHosts(c.DeckProviderAllowedHosts)) == 0 {
		return fmt.Errorf("DECK_PROVIDER_ALLOWED_HOSTS is required when DECK_PROVIDER_ENABLED=true")
	}
	return nil
}
