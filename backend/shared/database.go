package shared

import (
	"bytes"
	"context"
	"crypto/sha512"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/thienty1207/BWP-Staff/backend/config"
)

type migration struct {
	version     int64
	description string
	name        string
	contents    []byte
	checksum    []byte
}

func Connect(ctx context.Context, settings config.Config) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(settings.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse database URL: %w", err)
	}
	poolConfig.MaxConns = settings.DatabaseMaxConnections
	poolConfig.MinConns = settings.DatabaseMinConnections

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("create database pool: %w", err)
	}

	pingContext, cancel := context.WithTimeout(ctx, time.Duration(settings.DatabaseAcquireTimeoutSeconds)*time.Second)
	defer cancel()
	if err := pool.Ping(pingContext); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return pool, nil
}

func RunMigrations(ctx context.Context, pool *pgxpool.Pool, directory string) error {
	connection, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("acquire migration connection: %w", err)
	}
	defer connection.Release()

	if _, err := connection.Exec(ctx, `SELECT pg_advisory_lock(hashtext('bwp-sonasea:migrations'))`); err != nil {
		return fmt.Errorf("lock migrations: %w", err)
	}
	defer func() {
		_, _ = connection.Exec(context.Background(), `SELECT pg_advisory_unlock(hashtext('bwp-sonasea:migrations'))`)
	}()

	if _, err := connection.Exec(ctx, `
CREATE TABLE IF NOT EXISTS _sqlx_migrations (
    version BIGINT PRIMARY KEY,
    description TEXT NOT NULL,
    installed_on TIMESTAMPTZ NOT NULL DEFAULT now(),
    success BOOLEAN NOT NULL,
    checksum BYTEA NOT NULL,
    execution_time BIGINT NOT NULL
)`); err != nil {
		return fmt.Errorf("ensure migration ledger: %w", err)
	}

	migrations, err := readMigrations(directory)
	if err != nil {
		return err
	}
	for _, migration := range migrations {
		var storedChecksum []byte
		var success bool
		err := connection.QueryRow(ctx, `
SELECT checksum, success
FROM _sqlx_migrations
WHERE version = $1`, migration.version).Scan(&storedChecksum, &success)
		if err == nil {
			if !success {
				return fmt.Errorf("migration %s previously failed", migration.name)
			}
			if !bytes.Equal(storedChecksum, migration.checksum) {
				return fmt.Errorf("migration %s checksum mismatch", migration.name)
			}
			continue
		}
		if err != pgx.ErrNoRows {
			return fmt.Errorf("read migration %s status: %w", migration.name, err)
		}

		started := time.Now()
		transaction, err := connection.Begin(ctx)
		if err != nil {
			return fmt.Errorf("begin migration %s: %w", migration.name, err)
		}
		if _, err := transaction.Exec(ctx, string(migration.contents)); err != nil {
			_ = transaction.Rollback(ctx)
			return fmt.Errorf("execute migration %s: %w", migration.name, err)
		}
		_, err = transaction.Exec(ctx, `
INSERT INTO _sqlx_migrations (version, description, installed_on, success, checksum, execution_time)
VALUES ($1, $2, now(), TRUE, $3, $4)`,
			migration.version,
			migration.description,
			migration.checksum,
			time.Since(started).Microseconds(),
		)
		if err != nil {
			_ = transaction.Rollback(ctx)
			return fmt.Errorf("record migration %s: %w", migration.name, err)
		}
		if err := transaction.Commit(ctx); err != nil {
			return fmt.Errorf("commit migration %s: %w", migration.name, err)
		}
	}
	return nil
}

func readMigrations(directory string) ([]migration, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, fmt.Errorf("read migrations directory: %w", err)
	}

	migrations := make([]migration, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".sql" {
			continue
		}
		version, description, err := migrationName(entry.Name())
		if err != nil {
			return nil, err
		}
		contents, err := os.ReadFile(filepath.Join(directory, entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("read migration %s: %w", entry.Name(), err)
		}
		checksum := sha512.Sum384(contents)
		migrations = append(migrations, migration{
			version:     version,
			description: description,
			name:        entry.Name(),
			contents:    contents,
			checksum:    checksum[:],
		})
	}
	sort.Slice(migrations, func(i, j int) bool { return migrations[i].version < migrations[j].version })
	for i := 1; i < len(migrations); i++ {
		if migrations[i-1].version == migrations[i].version {
			return nil, fmt.Errorf("duplicate migration version %d", migrations[i].version)
		}
	}
	return migrations, nil
}

func migrationName(name string) (int64, string, error) {
	base := strings.TrimSuffix(name, filepath.Ext(name))
	parts := strings.SplitN(base, "_", 2)
	if len(parts) != 2 || parts[1] == "" {
		return 0, "", fmt.Errorf("invalid migration filename %s", name)
	}
	version, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || version <= 0 {
		return 0, "", fmt.Errorf("invalid migration version in %s", name)
	}
	return version, strings.ReplaceAll(parts[1], "_", " "), nil
}
