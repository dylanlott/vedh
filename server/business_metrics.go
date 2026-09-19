package server

import (
	"database/sql"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var businessMetricsOnce sync.Once

var (
	vedhSignupsTotal            *prometheus.CounterVec
	vedhLoginAttemptsTotal      *prometheus.CounterVec
	vedhGamesCreatedTotal       *prometheus.CounterVec
	vedhGameJoinAttemptsTotal   *prometheus.CounterVec
	vedhUsersTotal              prometheus.Gauge
	vedhDailyActiveUsersTotal   prometheus.Gauge
	vedhWeeklyActiveUsersTotal  prometheus.Gauge
	vedhMonthlyActiveUsersTotal prometheus.Gauge
)

func init() {
	registerBusinessMetrics()
}

func registerBusinessMetrics() {
	businessMetricsOnce.Do(func() {
		vedhSignupsTotal = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "vedh_signups_total",
				Help: "Total vEDH signup attempts by result.",
			},
			[]string{"result"},
		)
		vedhLoginAttemptsTotal = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "vedh_login_attempts_total",
				Help: "Total vEDH login attempts by result.",
			},
			[]string{"result"},
		)
		vedhGamesCreatedTotal = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "vedh_games_created_total",
				Help: "Total vEDH games created by format.",
			},
			[]string{"format"},
		)
		vedhGameJoinAttemptsTotal = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "vedh_game_join_attempts_total",
				Help: "Total vEDH game join attempts by result.",
			},
			[]string{"result"},
		)
		vedhUsersTotal = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "vedh_users_total",
				Help: "Total registered vEDH users.",
			},
		)
		vedhDailyActiveUsersTotal = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "vedh_daily_active_users_total",
				Help: "Distinct vEDH users with persisted activity in the last 24 hours.",
			},
		)
		vedhWeeklyActiveUsersTotal = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "vedh_weekly_active_users_total",
				Help: "Distinct vEDH users with persisted activity in the last 7 days.",
			},
		)
		vedhMonthlyActiveUsersTotal = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "vedh_monthly_active_users_total",
				Help: "Distinct vEDH users with persisted activity in the last 30 days.",
			},
		)

		prometheus.MustRegister(
			vedhSignupsTotal,
			vedhLoginAttemptsTotal,
			vedhGamesCreatedTotal,
			vedhGameJoinAttemptsTotal,
			vedhUsersTotal,
			vedhDailyActiveUsersTotal,
			vedhWeeklyActiveUsersTotal,
			vedhMonthlyActiveUsersTotal,
		)
		for _, result := range []string{"success", "validation_error", "conflict", "invalid_credentials", "error"} {
			vedhSignupsTotal.WithLabelValues(result)
			vedhLoginAttemptsTotal.WithLabelValues(result)
		}
		for _, format := range []string{"edh", "generic_duel", "unknown"} {
			vedhGamesCreatedTotal.WithLabelValues(format)
		}
		for _, result := range []string{"success", "invalid_input", "unauthorized", "not_found", "game_finished", "game_full", "already_in_game", "invalid_decklist", "error"} {
			vedhGameJoinAttemptsTotal.WithLabelValues(result)
		}
	})
}

func resetBusinessMetricsForTests() {
	registerBusinessMetrics()
	vedhSignupsTotal.Reset()
	vedhLoginAttemptsTotal.Reset()
	vedhGamesCreatedTotal.Reset()
	vedhGameJoinAttemptsTotal.Reset()
	vedhUsersTotal.Set(0)
	vedhDailyActiveUsersTotal.Set(0)
	vedhWeeklyActiveUsersTotal.Set(0)
	vedhMonthlyActiveUsersTotal.Set(0)
}

func recordVedhSignup(result string) {
	registerBusinessMetrics()
	vedhSignupsTotal.WithLabelValues(normalizeBusinessMetricLabel(result, "unknown")).Inc()
}

func recordVedhLoginAttempt(result string) {
	registerBusinessMetrics()
	vedhLoginAttemptsTotal.WithLabelValues(normalizeBusinessMetricLabel(result, "unknown")).Inc()
}

func recordVedhGameCreated(format string) {
	registerBusinessMetrics()
	vedhGamesCreatedTotal.WithLabelValues(normalizeBusinessMetricLabel(format, "unknown")).Inc()
}

func recordVedhGameJoinAttempt(result string) {
	registerBusinessMetrics()
	vedhGameJoinAttemptsTotal.WithLabelValues(normalizeBusinessMetricLabel(result, "unknown")).Inc()
}

func normalizeBusinessMetricLabel(value, fallback string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return fallback
	}
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			continue
		}
		return fallback
	}
	return value
}

func refreshVedhUserActivityMetrics(db *sql.DB) error {
	if db == nil {
		return nil
	}
	var totalUsersCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM "users"`).Scan(&totalUsersCount); err != nil {
		return err
	}
	var dailyCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM "users" WHERE last_active_at >= $1`, time.Now().UTC().Add(-24*time.Hour)).Scan(&dailyCount); err != nil {
		return err
	}
	var weeklyCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM "users" WHERE last_active_at >= $1`, time.Now().UTC().Add(-7*24*time.Hour)).Scan(&weeklyCount); err != nil {
		return err
	}
	var monthlyCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM "users" WHERE last_active_at >= $1`, time.Now().UTC().Add(-30*24*time.Hour)).Scan(&monthlyCount); err != nil {
		return err
	}
	vedhUsersTotal.Set(float64(totalUsersCount))
	vedhDailyActiveUsersTotal.Set(float64(dailyCount))
	vedhWeeklyActiveUsersTotal.Set(float64(weeklyCount))
	vedhMonthlyActiveUsersTotal.Set(float64(monthlyCount))
	return nil
}

func touchVedhUserActivity(db *sql.DB, userID string, now time.Time, minInterval time.Duration) (bool, error) {
	if db == nil || strings.TrimSpace(userID) == "" {
		return false, nil
	}
	var lastActiveAt sql.NullTime
	err := db.QueryRow(`SELECT last_active_at FROM "users" WHERE uuid = $1`, userID).Scan(&lastActiveAt)
	if err != nil {
		return false, err
	}
	if lastActiveAt.Valid && now.Sub(lastActiveAt.Time) < minInterval {
		return false, nil
	}
	if _, err := db.Exec(`UPDATE "users" SET last_active_at = $2 WHERE uuid = $1`, userID, now); err != nil {
		return false, err
	}
	return true, nil
}
