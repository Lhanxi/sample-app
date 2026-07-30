package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Lhanxi/sample-app/backend/internal/database"
	"github.com/jackc/pgx/v5/pgxpool"
)

const defaultMigrationsDirectory = "/app/migrations"

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "migration failed: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}

	migrationsDirectory := os.Getenv("MIGRATIONS_DIR")
	if migrationsDirectory == "" {
		migrationsDirectory = defaultMigrationsDirectory
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	pool, err := database.Open(ctx, databaseURL)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer pool.Close()

	return applyMigrations(ctx, pool, migrationsDirectory)
}

func applyMigrations(
	ctx context.Context,
	pool *pgxpool.Pool,
	migrationsDirectory string,
) error {
	if _, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`); err != nil {
		return fmt.Errorf("create migration history: %w", err)
	}

	entries, err := os.ReadDir(migrationsDirectory)
	if err != nil {
		return fmt.Errorf("read migrations directory: %w", err)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".up.sql") {
			continue
		}

		version := strings.TrimSuffix(entry.Name(), ".up.sql")
		applied, err := migrationApplied(ctx, pool, version)
		if err != nil {
			return err
		}
		if applied {
			logger.Info("migration already applied", "version", version)
			continue
		}

		migrationSQL, err := os.ReadFile(filepath.Join(migrationsDirectory, entry.Name()))
		if err != nil {
			return fmt.Errorf("read migration %s: %w", version, err)
		}

		tx, err := pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("begin migration %s: %w", version, err)
		}

		if _, err := tx.Exec(ctx, string(migrationSQL)); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("apply migration %s: %w", version, err)
		}
		if _, err := tx.Exec(
			ctx,
			"INSERT INTO schema_migrations (version) VALUES ($1)",
			version,
		); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("record migration %s: %w", version, err)
		}
		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("commit migration %s: %w", version, err)
		}

		logger.Info("migration applied", "version", version)
	}

	return nil
}

func migrationApplied(
	ctx context.Context,
	pool *pgxpool.Pool,
	version string,
) (bool, error) {
	var applied bool
	if err := pool.QueryRow(
		ctx,
		"SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE version = $1)",
		version,
	).Scan(&applied); err != nil {
		return false, fmt.Errorf("check migration %s: %w", version, err)
	}

	return applied, nil
}
