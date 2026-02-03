package pg

import (
	"context"
	"embed"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return err
	}
	var ups []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".up.sql") {
			ups = append(ups, e.Name())
		}
	}
	sort.Strings(ups)

	_, err = pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS payment_schema_migrations (version TEXT PRIMARY KEY)`)
	if err != nil {
		return err
	}

	for _, name := range ups {
		version := strings.TrimSuffix(name, ".up.sql")
		var applied string
		qerr := pool.QueryRow(ctx, `SELECT version FROM payment_schema_migrations WHERE version = $1`, version).Scan(&applied)
		if qerr == nil {
			continue
		}
		body, err := migrationsFS.ReadFile("migrations/" + name)
		if err != nil {
			return err
		}
		_, err = pool.Exec(ctx, string(body))
		if err != nil {
			return err
		}
		_, err = pool.Exec(ctx, `INSERT INTO payment_schema_migrations (version) VALUES ($1)`, version)
		if err != nil {
			return err
		}
	}
	return nil
}
