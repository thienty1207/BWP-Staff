package app

import "github.com/jackc/pgx/v5/pgxpool"

// AppState contains the concrete shared dependencies used by the HTTP app.
type AppState struct {
	DB *pgxpool.Pool
}
