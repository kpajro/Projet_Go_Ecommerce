// Package postgres contient l'implémentation de la couche de persistance
// basée sur PostgreSQL. Elle s'appuie uniquement sur database/sql : aucune ORM,
// les requêtes SQL sont écrites à la main.
package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	// Driver PostgreSQL compatible database/sql.
	_ "github.com/jackc/pgx/v5/stdlib"
)

// Open ouvre un pool de connexions vers PostgreSQL et configure le pool
// (nombre de connexions ouvertes, inactives et durée de vie maximale).
func Open(ctx context.Context, dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("postgres: ouverture impossible: %w", err)
	}
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)
	return db, nil
}

// Ping vérifie que la base répond. Utilisé par /healthz et au démarrage.
func Ping(ctx context.Context, db *sql.DB) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("postgres: base injoignable: %w", err)
	}
	return nil
}
