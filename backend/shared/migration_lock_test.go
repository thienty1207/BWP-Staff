package shared

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/thienty1207/BWP-Staff/backend/config"
)

func TestRunMigrationsTimesOutWhenLockHeld(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_TEST_URL")
	if databaseURL == "" {
		t.Skip("set DATABASE_TEST_URL to run the PostgreSQL migration lock contract")
	}
	t.Setenv("DATABASE_URL", databaseURL)
	t.Setenv("SEED_DEVELOPMENT_DATA", "false")

	settings, err := config.Load()
	if err != nil {
		t.Fatalf("load test database config: %v", err)
	}
	settings.DatabaseMaxConnections = 2
	settings.DatabaseMinConnections = 0

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pool, err := Connect(ctx, settings)
	if err != nil {
		t.Fatalf("connect test database: %v", err)
	}
	defer pool.Close()

	blocker, err := pool.Acquire(ctx)
	if err != nil {
		t.Fatalf("acquire lock blocker: %v", err)
	}
	defer blocker.Release()
	if _, err := blocker.Exec(ctx, `SELECT pg_advisory_lock(hashtext('bwp-sonasea:migrations'))`); err != nil {
		t.Fatalf("hold migration lock: %v", err)
	}

	err = RunMigrations(context.Background(), pool, filepath.Join("..", "migrations"), 100*time.Millisecond)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected migration lock timeout, got %v", err)
	}

	unlockContext, unlockCancel := context.WithTimeout(context.Background(), time.Second)
	defer unlockCancel()
	if _, unlockErr := blocker.Exec(unlockContext, `SELECT pg_advisory_unlock(hashtext('bwp-sonasea:migrations'))`); unlockErr != nil {
		t.Fatalf("release migration lock: %v", unlockErr)
	}
}
