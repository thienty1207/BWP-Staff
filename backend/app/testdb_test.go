package app

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func openAppTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	loadLocalTestEnv(t)
	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		t.Skip("set DATABASE_URL to run PostgreSQL-backed readiness tests")
	}
	if appEnv := strings.TrimSpace(os.Getenv("APP_ENV")); appEnv != "" && !strings.EqualFold(appEnv, "development") {
		t.Skipf("refusing PostgreSQL-backed test with APP_ENV=%q", appEnv)
	}

	parsedURL, err := url.Parse(databaseURL)
	if err != nil || !isLoopbackDatabaseURL(parsedURL) {
		t.Skip("refusing PostgreSQL-backed test unless DATABASE_URL points to a loopback PostgreSQL instance")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)

	schema := fmt.Sprintf("spec02_%d", time.Now().UnixNano())
	quotedSchema := pgx.Identifier{schema}.Sanitize()

	adminConnection, err := pgx.Connect(ctx, databaseURL)
	if err != nil {
		t.Skipf("PostgreSQL is unavailable for readiness tests: %v", err)
	}
	if _, err := adminConnection.Exec(ctx, "CREATE SCHEMA "+quotedSchema); err != nil {
		_ = adminConnection.Close(ctx)
		t.Skipf("cannot create isolated PostgreSQL readiness schema: %v", err)
	}
	if err := adminConnection.Close(ctx); err != nil {
		t.Fatalf("close PostgreSQL setup connection: %v", err)
	}

	query := parsedURL.Query()
	query.Del("options")
	encodedQuery := query.Encode()
	encodedOptions := strings.ReplaceAll(url.QueryEscape("-c search_path="+schema+",public"), "+", "%20")
	if encodedQuery != "" {
		encodedQuery += "&"
	}
	parsedURL.RawQuery = encodedQuery + "options=" + encodedOptions

	poolConfig, err := pgxpool.ParseConfig(parsedURL.String())
	if err != nil {
		dropAppTestSchema(t, ctx, databaseURL, quotedSchema)
		t.Fatalf("parse isolated PostgreSQL pool config: %v", err)
	}
	poolConfig.MaxConns = 1
	poolConfig.MinConns = 0
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		dropAppTestSchema(t, ctx, databaseURL, quotedSchema)
		t.Skipf("cannot create PostgreSQL readiness pool: %v", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		dropAppTestSchema(t, ctx, databaseURL, quotedSchema)
		t.Skipf("PostgreSQL is unavailable for readiness tests: %v", err)
	}

	t.Cleanup(func() {
		pool.Close()
		dropAppTestSchema(t, context.Background(), databaseURL, quotedSchema)
	})
	return pool
}

func loadLocalTestEnv(t *testing.T) {
	t.Helper()
	if strings.TrimSpace(os.Getenv("DATABASE_URL")) != "" {
		return
	}
	if err := godotenv.Load(filepath.Join("..", ".env")); err != nil && !os.IsNotExist(err) {
		t.Fatalf("load local backend/.env for PostgreSQL test: %v", err)
	}
}

func dropAppTestSchema(t *testing.T, ctx context.Context, databaseURL, quotedSchema string) {
	t.Helper()

	cleanupContext, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	connection, err := pgx.Connect(cleanupContext, databaseURL)
	if err != nil {
		t.Errorf("connect PostgreSQL cleanup connection: %v", err)
		return
	}
	defer connection.Close(cleanupContext)
	if _, err := connection.Exec(cleanupContext, "DROP SCHEMA "+quotedSchema+" CASCADE"); err != nil {
		t.Errorf("drop isolated PostgreSQL readiness schema: %v", err)
	}
}

func isLoopbackDatabaseURL(databaseURL *url.URL) bool {
	if databaseURL == nil || (databaseURL.Scheme != "postgres" && databaseURL.Scheme != "postgresql") {
		return false
	}
	host := strings.TrimSuffix(strings.ToLower(databaseURL.Hostname()), ".")
	if host == "localhost" || host == "127.0.0.1" || host == "::1" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}
