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
