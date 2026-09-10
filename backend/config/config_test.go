package config

import "testing"

func TestLoadRequiresOneDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://postgres:secret@127.0.0.1:5432/bwp-sonasea")
	t.Setenv("DATABASE_MAX_CONNECTIONS", "10")
	t.Setenv("DATABASE_MIN_CONNECTIONS", "1")
	t.Setenv("DATABASE_ACQUIRE_TIMEOUT_SECONDS", "5")
	t.Setenv("APP_ENV", "test")
	t.Setenv("SEED_DEVELOPMENT_DATA", "false")

	config, err := Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if config.DatabaseURL != "postgres://postgres:secret@127.0.0.1:5432/bwp-sonasea" {
		t.Fatalf("unexpected database URL: %s", config.DatabaseURL)
	}
	if config.DatabaseMaxConnections != 10 || config.DatabaseMinConnections != 1 {
		t.Fatalf("unexpected pool bounds: max=%d min=%d", config.DatabaseMaxConnections, config.DatabaseMinConnections)
	}
}

func TestLoadRequiresDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "")

	if _, err := Load(); err == nil {
		t.Fatal("expected missing DATABASE_URL error")
	}
}

func TestLoadRejectsRemoteDatabaseWithoutExplicitTLS(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://postgres:secret@db.example.com:5432/bwp-sonasea")
	t.Setenv("SEED_DEVELOPMENT_DATA", "false")

	if _, err := Load(); err == nil {
		t.Fatal("expected remote database URL without sslmode to be rejected")
	}
}

func TestLoadRejectsInsecureRemoteTLSMode(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://postgres:secret@db.example.com:5432/bwp-sonasea?sslmode=prefer")
	t.Setenv("SEED_DEVELOPMENT_DATA", "false")

	if _, err := Load(); err == nil {
		t.Fatal("expected remote database URL with sslmode=prefer to be rejected")
	}
}

func TestLoadAcceptsRemoteDatabaseWithExplicitTLS(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://postgres:secret@db.example.com:5432/bwp-sonasea?sslmode=verify-full")
	t.Setenv("SEED_DEVELOPMENT_DATA", "false")

	if _, err := Load(); err != nil {
		t.Fatalf("expected remote database URL with sslmode=verify-full to be accepted: %v", err)
	}
}

func TestLoadRejectsPoolAboveSafeLimit(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://postgres@127.0.0.1:5432/bwp-sonasea")
	t.Setenv("DATABASE_MAX_CONNECTIONS", "101")
	t.Setenv("DATABASE_MIN_CONNECTIONS", "1")
	t.Setenv("SEED_DEVELOPMENT_DATA", "false")

	if _, err := Load(); err == nil {
		t.Fatal("expected database pool above the safe limit to be rejected")
	}
}

func TestLoadRejectsAcquireTimeoutAboveSafeLimit(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://postgres@127.0.0.1:5432/bwp-sonasea")
	t.Setenv("DATABASE_ACQUIRE_TIMEOUT_SECONDS", "61")
	t.Setenv("SEED_DEVELOPMENT_DATA", "false")

	if _, err := Load(); err == nil {
		t.Fatal("expected database acquire timeout above the safe limit to be rejected")
	}
}
