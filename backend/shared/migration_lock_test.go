package shared_test

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/thienty1207/BWP-Staff/backend/shared"
)

func TestRunMigrationsTimesOutWhenLockHeld(t *testing.T) {
	pool, ctx := openSPEC01Pool(t)

	blocker, err := pool.Acquire(ctx)
	if err != nil {
		t.Fatalf("acquire lock blocker: %v", err)
	}
	defer blocker.Release()
	if _, err := blocker.Exec(ctx, `SELECT pg_advisory_lock(hashtext('bwp-sonasea:migrations'))`); err != nil {
		t.Fatalf("hold migration lock: %v", err)
	}

	err = shared.RunMigrations(context.Background(), pool, filepath.Join("..", "migrations"), 100*time.Millisecond)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected migration lock timeout, got %v", err)
	}

	unlockContext, unlockCancel := context.WithTimeout(context.Background(), time.Second)
	defer unlockCancel()
	if _, unlockErr := blocker.Exec(unlockContext, `SELECT pg_advisory_unlock(hashtext('bwp-sonasea:migrations'))`); unlockErr != nil {
		t.Fatalf("release migration lock: %v", unlockErr)
	}
}
