package persistence

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/zeebo/errs"
)

var (
	// DefaultPGURl is exported for testing purposes
	DefaultPGURL = "postgres://edhgo:edhgodev@localhost:5432/edhgo?sslmode=disable"
	// Directory used for loading migrations
	defaultMigrationsDir = "./persistence/migrations/"
)

// NewDB returns a migrated app database or an error. If it is passed an
// empty string, it will attempt to connect to the default testing DB
func NewDB(dbURL string) (*sql.DB, error) {
	pg, err := NewPostgres(defaultMigrationsDir, dbURL)
	if err != nil {
		slog.Default().Error("failed to get postgres", "err", err)
		return nil, errs.Wrap(err)
	}
	if err := pg.Ping(); err != nil {
		return nil, errs.Wrap(err)
	}
	return pg, nil
}

// NewPostgres returns a migrated sql.DB with a Postgres database connection
// migdir is the relative path to the migrations directory.
func NewPostgres(migdir string, dbURL string) (*sql.DB, error) {
	slog.Default().Info("opening PostgreSQL database connection")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		slog.Default().Error("failed to open postgres connection", "err", err)
		return nil, errs.Wrap(err)
	}
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		slog.Default().Error("failed to create postgres migrate instance", "err", err)
		return nil, err
	}
	formattedMigrationsDir := fmt.Sprintf("file://%s", migdir)
	m, err := migrate.NewWithDatabaseInstance(formattedMigrationsDir, "postgres", driver)
	if err != nil {
		slog.Default().Error("failed to create migration instance", "err", err)
		return nil, errs.Wrap(err)
	}
	err = m.Up()
	if err != nil {
		if err.Error() == "no change" {
			v, dirty, err := m.Version()
			if err != nil {
				return nil, fmt.Errorf("failed to get migration version: %w", err)
			}
			if dirty {
				slog.Default().Warn("database migration state is dirty", "version", v)
			} else {
				slog.Default().Info("no migration changes detected", "latest_version", v)
			}
			return db, nil
		}
		// should we fail here? regardless, we should not be silent
		slog.Default().Error("failed to run migrations", "err", err)
		slog.Default().Warn("attempting migration rollback")
		return nil, m.Down()
	}

	slog.Default().Info("database created")
	return db, err
}

// ForceCleanMigrations clears a dirty migration state by forcing the current version.
// Intended for test setups that reuse a shared local database.
func ForceCleanMigrations(migdir string, dbURL string) error {
	slog.Default().Info("opening PostgreSQL database connection")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		slog.Default().Error("failed to open postgres connection", "err", err)
		return errs.Wrap(err)
	}
	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		slog.Default().Error("failed to create postgres migrate instance", "err", err)
		return err
	}
	formattedMigrationsDir := fmt.Sprintf("file://%s", migdir)
	m, err := migrate.NewWithDatabaseInstance(formattedMigrationsDir, "postgres", driver)
	if err != nil {
		slog.Default().Error("failed to create migration instance", "err", err)
		return errs.Wrap(err)
	}
	if err := clearDirtyMigration(m); err != nil {
		slog.Default().Error("failed to clear dirty migration state", "err", err)
		return errs.Wrap(err)
	}
	return nil
}

// MigrateDown rolls back all migrations for the given migrations directory.
// It is intended for test cleanup and will ignore "no change" errors.
func MigrateDown(migdir string, dbURL string) error {
	slog.Default().Info("opening PostgreSQL database connection")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		slog.Default().Error("failed to open postgres connection", "err", err)
		return errs.Wrap(err)
	}
	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		slog.Default().Error("failed to create postgres migrate instance", "err", err)
		return err
	}
	formattedMigrationsDir := fmt.Sprintf("file://%s", migdir)
	m, err := migrate.NewWithDatabaseInstance(formattedMigrationsDir, "postgres", driver)
	if err != nil {
		slog.Default().Error("failed to create migration instance", "err", err)
		return errs.Wrap(err)
	}
	if err := clearDirtyMigration(m); err != nil {
		slog.Default().Error("failed to clear dirty migration state", "err", err)
		return errs.Wrap(err)
	}
	if err := m.Down(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) || err.Error() == "no change" {
			return nil
		}
		slog.Default().Error("failed to run migrations down", "err", err)
		return errs.Wrap(err)
	}
	return nil
}

func clearDirtyMigration(m *migrate.Migrate) error {
	version, dirty, err := m.Version()
	if err != nil {
		if errors.Is(err, migrate.ErrNilVersion) {
			return nil
		}
		return err
	}
	if !dirty {
		return nil
	}
	return m.Force(int(version))
}
