package storage

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "modernc.org/sqlite"

	"github.com/agentplexus/multi-agent-designer/internal/ent"
)

const (
	// DefaultPath is the default SQLite database path.
	DefaultPath = "data/specui.db"
)

// NewClient creates a new Ent client with SQLite database.
// The database file is created at the specified path, or DefaultPath if empty.
func NewClient(ctx context.Context, dbPath string) (*ent.Client, error) {
	if dbPath == "" {
		dbPath = DefaultPath
	}

	// Ensure data directory exists
	dir := filepath.Dir(dbPath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("creating data directory: %w", err)
		}
	}

	// Open SQLite database with foreign keys enabled
	dsn := fmt.Sprintf("file:%s?_pragma=foreign_keys(1)", dbPath)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("opening sqlite: %w", err)
	}

	// Enable WAL mode for better concurrency
	if _, err := db.ExecContext(ctx, "PRAGMA journal_mode=WAL;"); err != nil {
		if closeErr := db.Close(); closeErr != nil {
			log.Printf("error closing database after WAL failure: %v", closeErr)
		}
		return nil, fmt.Errorf("enabling WAL mode: %w", err)
	}

	// Create Ent client
	drv := entsql.OpenDB(dialect.SQLite, db)
	client := ent.NewClient(ent.Driver(drv))

	// Run auto-migration
	if err := client.Schema.Create(ctx); err != nil {
		if closeErr := client.Close(); closeErr != nil {
			log.Printf("error closing client after migration failure: %v", closeErr)
		}
		return nil, fmt.Errorf("running migrations: %w", err)
	}

	return client, nil
}
