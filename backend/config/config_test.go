package config

import (
	"strings"
	"testing"
)

func setBaseConfigEnvironment(t *testing.T) {
	t.Helper()
	t.Setenv("DATABASE_URL", "postgres://postgres:secret@127.0.0.1:5432/bwp-sonasea")
	t.Setenv("DATABASE_MAX_CONNECTIONS", "10")
	t.Setenv("DATABASE_MIN_CONNECTIONS", "1")
	t.Setenv("DATABASE_ACQUIRE_TIMEOUT_SECONDS", "5")
	t.Setenv("APP_ENV", "development")
	t.Setenv("SEED_DEVELOPMENT_DATA", "false")
	t.Setenv("FRONTEND_ORIGIN", "")
	t.Setenv("BACKEND_BIND_ADDRESS", "")
	t.Setenv("BACKEND_SHUTDOWN_TIMEOUT_SECONDS", "")
	t.Setenv("AUTH_SESSION_TTL_HOURS", "")
}

func TestLoadRequiresOneDatabaseURL(t *testing.T) {
	setBaseConfigEnvironment(t)

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
	setBaseConfigEnvironment(t)
	t.Setenv("DATABASE_URL", "")

	if _, err := Load(); err == nil {
		t.Fatal("expected missing DATABASE_URL error")
	}
}

func TestLoadRejectsRemoteDatabaseWithoutExplicitTLS(t *testing.T) {
	setBaseConfigEnvironment(t)
	t.Setenv("DATABASE_URL", "postgres://postgres:secret@db.example.com:5432/bwp-sonasea")

	if _, err := Load(); err == nil {
		t.Fatal("expected remote database URL without sslmode to be rejected")
	}
}

func TestLoadRejectsInsecureRemoteTLSMode(t *testing.T) {
	setBaseConfigEnvironment(t)
	t.Setenv("DATABASE_URL", "postgres://postgres:secret@db.example.com:5432/bwp-sonasea?sslmode=prefer")

	if _, err := Load(); err == nil {
		t.Fatal("expected remote database URL with sslmode=prefer to be rejected")
	}
}

func TestLoadAcceptsRemoteDatabaseWithExplicitTLS(t *testing.T) {
	setBaseConfigEnvironment(t)
	t.Setenv("DATABASE_URL", "postgres://postgres:secret@db.example.com:5432/bwp-sonasea?sslmode=verify-full")

	if _, err := Load(); err != nil {
		t.Fatalf("expected remote database URL with sslmode=verify-full to be accepted: %v", err)
	}
}

func TestLoadRejectsPoolAboveSafeLimit(t *testing.T) {
	setBaseConfigEnvironment(t)
	t.Setenv("DATABASE_URL", "postgres://postgres@127.0.0.1:5432/bwp-sonasea")
	t.Setenv("DATABASE_MAX_CONNECTIONS", "101")

	if _, err := Load(); err == nil {
		t.Fatal("expected database pool above the safe limit to be rejected")
	}
}

func TestLoadRejectsAcquireTimeoutAboveSafeLimit(t *testing.T) {
	setBaseConfigEnvironment(t)
	t.Setenv("DATABASE_URL", "postgres://postgres@127.0.0.1:5432/bwp-sonasea")
	t.Setenv("DATABASE_ACQUIRE_TIMEOUT_SECONDS", "61")

	if _, err := Load(); err == nil {
		t.Fatal("expected database acquire timeout above the safe limit to be rejected")
	}
}

func TestLoadAppliesDevelopmentHTTPDefaults(t *testing.T) {
	setBaseConfigEnvironment(t)

	settings, err := Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if settings.BackendBindAddress != "127.0.0.1:3000" {
		t.Fatalf("unexpected bind address: %q", settings.BackendBindAddress)
	}
	if settings.FrontendOrigin != "http://localhost:5173" {
		t.Fatalf("unexpected frontend origin: %q", settings.FrontendOrigin)
	}
	if settings.BackendShutdownTimeoutSeconds != 10 {
		t.Fatalf("unexpected shutdown timeout: %d", settings.BackendShutdownTimeoutSeconds)
	}
	if settings.AuthSessionTTLHours != 12 {
		t.Fatalf("unexpected auth session TTL: %d", settings.AuthSessionTTLHours)
	}
}

func TestLoadAppliesHTTPOverrides(t *testing.T) {
	setBaseConfigEnvironment(t)
	t.Setenv("BACKEND_BIND_ADDRESS", ":3300")
	t.Setenv("FRONTEND_ORIGIN", "https://staff.example.com")
	t.Setenv("BACKEND_SHUTDOWN_TIMEOUT_SECONDS", "30")

	settings, err := Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if settings.BackendBindAddress != ":3300" {
		t.Fatalf("unexpected bind address: %q", settings.BackendBindAddress)
	}
	if settings.FrontendOrigin != "https://staff.example.com" {
		t.Fatalf("unexpected frontend origin: %q", settings.FrontendOrigin)
	}
	if settings.BackendShutdownTimeoutSeconds != 30 {
		t.Fatalf("unexpected shutdown timeout: %d", settings.BackendShutdownTimeoutSeconds)
	}
}

func TestLoadAppliesAuthSessionTTLOverride(t *testing.T) {
	setBaseConfigEnvironment(t)
	t.Setenv("AUTH_SESSION_TTL_HOURS", "36")

	settings, err := Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if settings.AuthSessionTTLHours != 36 {
		t.Fatalf("unexpected auth session TTL: %d", settings.AuthSessionTTLHours)
	}
}

func TestLoadRejectsInvalidAuthSessionTTL(t *testing.T) {
	for _, value := range []string{"0", "-1", "721"} {
		t.Run(value, func(t *testing.T) {
			setBaseConfigEnvironment(t)
			t.Setenv("AUTH_SESSION_TTL_HOURS", value)

			if _, err := Load(); err == nil || !strings.Contains(err.Error(), "AUTH_SESSION_TTL_HOURS") {
				t.Fatalf("expected invalid auth session TTL error, got %v", err)
			}
		})
	}
}

func TestLoadRequiresFrontendOriginOutsideDevelopment(t *testing.T) {
	setBaseConfigEnvironment(t)
	t.Setenv("APP_ENV", "production")

	_, err := Load()
	if err == nil || !strings.Contains(err.Error(), "FRONTEND_ORIGIN") {
		t.Fatalf("expected missing production frontend origin error, got %v", err)
	}
}

func TestLoadRejectsWildcardFrontendOrigin(t *testing.T) {
	setBaseConfigEnvironment(t)
	t.Setenv("FRONTEND_ORIGIN", "*")

	_, err := Load()
	if err == nil || !strings.Contains(err.Error(), "FRONTEND_ORIGIN") {
		t.Fatalf("expected wildcard frontend origin error, got %v", err)
	}
}

func TestLoadRejectsInvalidFrontendOrigin(t *testing.T) {
	invalidOrigins := []string{
		"ftp://staff.example.com",
		"https://staff.example.com/path",
		"https://staff.example.com?tenant=1",
		"https://staff.example.com?",
		"https://staff.example.com#section",
		"https://staff.example.com#",
	}

	for _, origin := range invalidOrigins {
		t.Run(origin, func(t *testing.T) {
			setBaseConfigEnvironment(t)
			t.Setenv("FRONTEND_ORIGIN", origin)

			_, err := Load()
			if err == nil || !strings.Contains(err.Error(), "FRONTEND_ORIGIN") {
				t.Fatalf("expected invalid frontend origin error, got %v", err)
			}
		})
	}
}

func TestLoadRejectsInvalidBindAddress(t *testing.T) {
	invalidAddresses := []string{"3000", "http://127.0.0.1:3000", "*:3000", ":0", ":65536", ":not-a-port"}

	for _, address := range invalidAddresses {
		t.Run(address, func(t *testing.T) {
			setBaseConfigEnvironment(t)
			t.Setenv("BACKEND_BIND_ADDRESS", address)

			_, err := Load()
			if err == nil || !strings.Contains(err.Error(), "BACKEND_BIND_ADDRESS") {
				t.Fatalf("expected invalid bind address error, got %v", err)
			}
		})
	}
}

func TestLoadRejectsInvalidShutdownTimeout(t *testing.T) {
	invalidTimeouts := []string{"0", "-1", "61"}

	for _, timeout := range invalidTimeouts {
		t.Run(timeout, func(t *testing.T) {
			setBaseConfigEnvironment(t)
			t.Setenv("BACKEND_SHUTDOWN_TIMEOUT_SECONDS", timeout)

			_, err := Load()
			if err == nil || !strings.Contains(err.Error(), "BACKEND_SHUTDOWN_TIMEOUT_SECONDS") {
				t.Fatalf("expected invalid shutdown timeout error, got %v", err)
			}
		})
	}
}
