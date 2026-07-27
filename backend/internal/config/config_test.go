package config

import (
	"strings"
	"testing"
	"time"
)

func TestLoadUsesDefaultValues(t *testing.T) {
	t.Setenv("APP_ENV", "")
	t.Setenv("PORT", "")
	t.Setenv("READ_TIMEOUT", "")
	t.Setenv("WRITE_TIMEOUT", "")
	t.Setenv("IDLE_TIMEOUT", "")
	t.Setenv("SHUTDOWN_TIMEOUT", "")
	t.Setenv("DATABASE_URL", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned an unexpected error: %v", err)
	}

	if cfg.Environment != "development" {
		t.Errorf(
			"Environment = %q; want %q",
			cfg.Environment,
			"development",
		)
	}

	if cfg.Port != 8080 {
		t.Errorf("Port = %d; want %d", cfg.Port, 8080)
	}

	if cfg.ReadTimeout != 10*time.Second {
		t.Errorf(
			"ReadTimeout = %v; want %v",
			cfg.ReadTimeout,
			10*time.Second,
		)
	}

	if cfg.WriteTimeout != 15*time.Second {
		t.Errorf(
			"WriteTimeout = %v; want %v",
			cfg.WriteTimeout,
			15*time.Second,
		)
	}

	if cfg.IdleTimeout != 60*time.Second {
		t.Errorf(
			"IdleTimeout = %v; want %v",
			cfg.IdleTimeout,
			60*time.Second,
		)
	}

	if cfg.ShutdownTimeout != 10*time.Second {
		t.Errorf(
			"ShutdownTimeout = %v; want %v",
			cfg.ShutdownTimeout,
			10*time.Second,
		)
	}

	const defaultDatabaseURL = "postgres://sample:sample@localhost:5432/sample?sslmode=disable"
	if cfg.DatabaseURL != defaultDatabaseURL {
		t.Errorf("DatabaseURL = %q; want %q", cfg.DatabaseURL, defaultDatabaseURL)
	}
}

func TestLoadUsesEnvironmentValues(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("PORT", "9090")
	t.Setenv("READ_TIMEOUT", "5s")
	t.Setenv("WRITE_TIMEOUT", "7s")
	t.Setenv("IDLE_TIMEOUT", "30s")
	t.Setenv("SHUTDOWN_TIMEOUT", "20s")
	t.Setenv("DATABASE_URL", "postgres://test:test@database:5432/test")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned an unexpected error: %v", err)
	}

	if cfg.Environment != "production" {
		t.Errorf(
			"Environment = %q; want %q",
			cfg.Environment,
			"production",
		)
	}

	if cfg.Port != 9090 {
		t.Errorf("Port = %d; want %d", cfg.Port, 9090)
	}

	if cfg.ReadTimeout != 5*time.Second {
		t.Errorf(
			"ReadTimeout = %v; want %v",
			cfg.ReadTimeout,
			5*time.Second,
		)
	}

	if cfg.WriteTimeout != 7*time.Second {
		t.Errorf(
			"WriteTimeout = %v; want %v",
			cfg.WriteTimeout,
			7*time.Second,
		)
	}

	if cfg.IdleTimeout != 30*time.Second {
		t.Errorf(
			"IdleTimeout = %v; want %v",
			cfg.IdleTimeout,
			30*time.Second,
		)
	}

	if cfg.ShutdownTimeout != 20*time.Second {
		t.Errorf(
			"ShutdownTimeout = %v; want %v",
			cfg.ShutdownTimeout,
			20*time.Second,
		)
	}

	if cfg.DatabaseURL != "postgres://test:test@database:5432/test" {
		t.Errorf(
			"DatabaseURL = %q; want %q",
			cfg.DatabaseURL,
			"postgres://test:test@database:5432/test",
		)
	}
}

func TestLoadReturnsErrorForInvalidPort(t *testing.T) {
	t.Setenv("PORT", "not-a-number")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() error = nil; want an error")
	}

	if !strings.Contains(err.Error(), "PORT must be an integer") {
		t.Errorf(
			"Load() error = %q; want error containing %q",
			err.Error(),
			"PORT must be an integer",
		)
	}
}

func TestLoadReturnsErrorForInvalidDuration(t *testing.T) {
	t.Setenv("READ_TIMEOUT", "invalid")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() error = nil; want an error")
	}

	if !strings.Contains(err.Error(), "READ_TIMEOUT must be a valid duration") {
		t.Errorf(
			"Load() error = %q; want error containing %q",
			err.Error(),
			"READ_TIMEOUT must be a valid duration",
		)
	}
}
