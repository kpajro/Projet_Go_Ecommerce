package postgres

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"sort"
)

// migrationsFS contient les fichiers SQL embarqués dans le binaire.
// go:embed impose que les fichiers soient dans le package (ou un sous-dossier).
//
//go:embed migrations/*.sql
var migrationsFS embed.FS

// Migrate applique les migrations non encore jouées, dans l'ordre lexicographique.
// Chaque migration est jouée dans sa propre transaction : soit elle passe en
// entier, soit rien n'est écrit. Les versions déjà appliquées sont ignorées,
// ce qui rend l'opération idempotente et sûre au redémarrage du serveur.
//
// Le suivi se fait dans la table schema_migrations.
func Migrate(ctx context.Context, db *sql.DB) error {
	const createTable = `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version    TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)`

	if _, err := db.ExecContext(ctx, createTable); err != nil {
		return fmt.Errorf("postgres: création de schema_migrations impossible: %w", err)
	}

	applied, err := appliedVersions(ctx, db)
	if err != nil {
		return err
	}

	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("postgres: lecture des migrations impossible: %w", err)
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		version := entry.Name()
		if applied[version] {
			continue
		}

		content, err := migrationsFS.ReadFile("migrations/" + version)
		if err != nil {
			return fmt.Errorf("postgres: lecture de la migration %s impossible: %w", version, err)
		}

		if err := applyOne(ctx, db, version, string(content)); err != nil {
			return err
		}
	}
	return nil
}

// applyOne joue une migration puis enregistre sa version.
func applyOne(ctx context.Context, db *sql.DB, version, content string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("postgres: démarrage de la transaction impossible: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, content); err != nil {
		return fmt.Errorf("postgres: migration %s en échec: %w", version, err)
	}
	if _, err := tx.ExecContext(ctx,
		"INSERT INTO schema_migrations (version) VALUES ($1)", version); err != nil {
		return fmt.Errorf("postgres: enregistrement de la migration %s impossible: %w", version, err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("postgres: validation de la migration %s impossible: %w", version, err)
	}
	return nil
}

// appliedVersions retourne l'ensemble des migrations déjà appliquées.
func appliedVersions(ctx context.Context, db *sql.DB) (map[string]bool, error) {
	rows, err := db.QueryContext(ctx, "SELECT version FROM schema_migrations")
	if err != nil {
		return nil, fmt.Errorf("postgres: lecture de schema_migrations impossible: %w", err)
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, fmt.Errorf("postgres: lecture de schema_migrations impossible: %w", err)
		}
		applied[version] = true
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("postgres: lecture de schema_migrations impossible: %w", err)
	}
	return applied, nil
}
