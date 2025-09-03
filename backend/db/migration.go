package db

import (
	"embed"
	"fmt"

	"github.com/pressly/goose/v3"
)

//go:embed migrations/sqlite/*.sql
var EmbedMigrations embed.FS

// ApplyEmbeddedDbSchema applies database migrations using existing db connection
// version: -2 = apply all pending migrations
// version: -1 = apply single migration
// version: n = apply to specific version number
func ApplyEmbeddedDbSchema(version int64) error {
	goose.SetBaseFS(EmbedMigrations)

	if err := goose.SetDialect("sqlite3"); err != nil {
		return fmt.Errorf("failed to set dialect: %w", err)
	}

	// Use the existing writerDb instance from common.go
	writerMutex.Lock()
	defer writerMutex.Unlock()

	switch version {
	case -2:
		// Apply all pending migrations
		if err := goose.Up(writerDb.DB, "migrations/sqlite"); err != nil {
			return fmt.Errorf("failed to apply all migrations: %w", err)
		}
		logger.Info("Applied all pending migrations")
	case -1:
		// Apply single migration
		if err := goose.UpByOne(writerDb.DB, "migrations/sqlite"); err != nil {
			return fmt.Errorf("failed to apply single migration: %w", err)
		}
		logger.Info("Applied single migration")
	default:
		// Apply to specific version
		if err := goose.UpTo(writerDb.DB, "migrations/sqlite", version); err != nil {
			return fmt.Errorf("failed to apply to version %d: %w", version, err)
		}
		logger.Infof("Applied migrations to version %d", version)
	}

	return nil
}

// GetDbVersion returns the current database schema version
func GetDbVersion() (int64, error) {
	version, err := goose.GetDBVersion(ReaderDb.DB)
	if err != nil {
		return 0, fmt.Errorf("failed to get version: %w", err)
	}

	return version, nil
}

// RunMigrations applies all pending migrations
func RunMigrations() error {
	logger.Info("Running database migrations...")

	// Check current version
	currentVersion, err := GetDbVersion()
	if err != nil {
		logger.WithError(err).Warn("Failed to get current db version, assuming new database")
		currentVersion = 0
	}

	logger.Infof("Current database version: %d", currentVersion)

	// Apply all pending migrations
	if err := ApplyEmbeddedDbSchema(-2); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Get new version
	newVersion, err := GetDbVersion()
	if err != nil {
		return fmt.Errorf("failed to get new version: %w", err)
	}

	if newVersion > currentVersion {
		logger.Infof("Database migrated from version %d to %d", currentVersion, newVersion)
	} else {
		logger.Info("Database is up to date")
	}

	return nil
}
