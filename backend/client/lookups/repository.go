package lookups

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (repository *Repository) ListDepartments(ctx context.Context) ([]Department, error) {
	if repository == nil || repository.pool == nil {
		return nil, fmt.Errorf("lookup repository is not configured")
	}

	rows, err := repository.pool.Query(ctx, `
SELECT id, code, name
FROM departments
WHERE is_active = TRUE
ORDER BY name ASC, id ASC`)
	if err != nil {
		return nil, fmt.Errorf("query departments: %w", err)
	}
	defer rows.Close()

	departments := make([]Department, 0)
	for rows.Next() {
		var department Department
		if err := rows.Scan(&department.ID, &department.Code, &department.Name); err != nil {
			return nil, fmt.Errorf("scan department: %w", err)
		}
		departments = append(departments, department)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate departments: %w", err)
	}
	return departments, nil
}

func (repository *Repository) ListLocations(ctx context.Context, query LocationQuery) ([]Location, error) {
	if repository == nil || repository.pool == nil {
		return nil, fmt.Errorf("lookup repository is not configured")
	}

	rows, err := repository.pool.Query(ctx, `
SELECT id, code, name
FROM locations
WHERE is_active = TRUE
  AND ($1 = '' OR strpos(LOWER(name), LOWER($1)) > 0)
ORDER BY
    CASE
        WHEN $1 = '' THEN 4
        WHEN LOWER(name) = LOWER($1) THEN 0
        WHEN strpos(LOWER(name), LOWER($1)) = 1 THEN 1
        WHEN EXISTS (
            SELECT 1
            FROM regexp_split_to_table(LOWER(name), '[^[:alnum:]]+') AS token
            WHERE token <> '' AND strpos(token, LOWER($1)) = 1
        ) THEN 2
        ELSE 3
    END ASC,
    name ASC,
    id ASC
LIMIT NULLIF($2::int, 0)`, query.Search, effectiveLocationLimit(query))
	if err != nil {
		return nil, fmt.Errorf("query locations: %w", err)
	}
	defer rows.Close()

	locations := make([]Location, 0)
	for rows.Next() {
		var location Location
		if err := rows.Scan(&location.ID, &location.Code, &location.Name); err != nil {
			return nil, fmt.Errorf("scan location: %w", err)
		}
		locations = append(locations, location)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate locations: %w", err)
	}
	return locations, nil
}

func effectiveLocationLimit(query LocationQuery) int {
	if !query.HasLimit {
		return 0
	}
	return query.Limit
}
