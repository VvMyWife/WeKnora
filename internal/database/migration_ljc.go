package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4"
	migratedb "github.com/golang-migrate/migrate/v4/database"
	pgmigrate "github.com/golang-migrate/migrate/v4/database/postgres"
	sqlite3migrate "github.com/golang-migrate/migrate/v4/database/sqlite3"
)

const ljcMigrationTable = "ljc_schema_migrations"

type ljcMigrationLayout struct {
	sqlite             bool
	base, legacyLast   int
	path, officialPath string
	groups             [][]string
}

func ljcLayout(dsn string, opts MigrationOptions) ljcMigrationLayout {
	if opts.SQLiteDBPath != "" || strings.HasPrefix(dsn, "sqlite3://") {
		return ljcMigrationLayout{true, 12, 15, "migrations/ljc/sqlite", "migrations/sqlite", [][]string{
			{"archive_mapping_batches", "archive_mapping_items", "archive_mapping_versions", "archive_mapping_links"},
			{"external_chunk_batches", "external_chunk_records", "external_chunk_memberships"},
			{"message_citation_manifests", "message_citation_evidence"},
		}}
	}
	return ljcMigrationLayout{false, 90, 94, "migrations/ljc/postgres", "migrations/versioned", [][]string{
		{"archive_mapping_batches", "archive_mapping_items", "archive_mapping_versions", "archive_mapping_links"},
		{"enrichment_retry_settings", "enrichment_recoveries", "enrichment_recovery_sources"},
		{"external_chunk_batches", "external_chunk_records", "external_chunk_memberships"},
		{"message_citation_manifests", "message_citation_evidence"},
	}}
}

func ljcOpenDB(dsn string, opts MigrationOptions) (*sql.DB, error) {
	if opts.SQLiteDBPath != "" {
		return sql.Open("sqlite3", opts.SQLiteDBPath)
	}
	if strings.HasPrefix(dsn, "sqlite3://") {
		return sql.Open("sqlite3", strings.TrimPrefix(dsn, "sqlite3://"))
	}
	return sql.Open("postgres", dsn)
}

func ljcTableExists(ctx context.Context, tx *sql.Tx, sqlite bool, name string) (bool, error) {
	var n int
	query := "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema=current_schema() AND table_name=$1"
	if sqlite {
		query = "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?"
	}
	err := tx.QueryRowContext(ctx, query, name).Scan(&n)
	return n > 0, err
}

// prepareLJCLegacyBridge recognizes only the frozen pre-fork layout. It never
// guesses a version, runs a down migration, or changes application data.
func prepareLJCLegacyBridge(dsn string, opts MigrationOptions) error {
	return prepareLJCLegacyBridgeWithLayout(dsn, opts, ljcLayout(dsn, opts))
}

func prepareLJCLegacyBridgeWithLayout(dsn string, opts MigrationOptions, layout ljcMigrationLayout) error {
	if _, err := os.Stat(layout.path); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}
	db, err := ljcOpenDB(dsn, opts)
	if err != nil {
		return err
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	exists, err := ljcTableExists(ctx, tx, layout.sqlite, "schema_migrations")
	if err != nil || !exists {
		return err
	}
	query := "SELECT version,dirty FROM schema_migrations"
	if !layout.sqlite {
		query += " FOR UPDATE"
	}
	var version int
	var dirty bool
	err = tx.QueryRowContext(ctx, query).Scan(&version, &dirty)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return err
	}
	if dirty {
		return fmt.Errorf("LJC bridge refuses dirty official migration state %d", version)
	}
	exists, err = ljcTableExists(ctx, tx, layout.sqlite, ljcMigrationTable)
	if err != nil {
		return err
	}
	if exists {
		var customVersion int
		var customDirty bool
		err = tx.QueryRowContext(ctx, "SELECT version,dirty FROM "+ljcMigrationTable).Scan(&customVersion, &customDirty)
		if err == nil && customDirty {
			return fmt.Errorf("LJC migration state %d is dirty", customVersion)
		}
		if err != nil && err != sql.ErrNoRows {
			return err
		}
		if version > layout.base && version <= layout.legacyLast && err == sql.ErrNoRows {
			return fmt.Errorf("ambiguous empty LJC ledger with official version %d", version)
		}
		return nil
	}
	if version <= layout.base {
		return nil
	}
	// A genuine upstream-only database can acquire the personal ledger on its
	// first custom start. Require the official MCP marker and absence of every
	// custom table; a partially missing custom ledger must never be guessed.
	var officialMarker int
	markerQuery := "SELECT COUNT(*) FROM information_schema.columns WHERE table_schema=current_schema() AND table_name='mcp_tool_approvals' AND column_name='enabled'"
	if layout.sqlite {
		markerQuery = "SELECT COUNT(*) FROM pragma_table_info('mcp_tool_approvals') WHERE name='enabled'"
	}
	if err := tx.QueryRowContext(ctx, markerQuery).Scan(&officialMarker); err != nil {
		return err
	}
	if officialMarker == 1 {
		for _, group := range layout.groups {
			for _, table := range group {
				present, err := ljcTableExists(ctx, tx, layout.sqlite, table)
				if err != nil {
					return err
				}
				if present {
					return fmt.Errorf("official marker and personal tables without LJC ledger: refusing ambiguous state")
				}
			}
		}
		return nil
	}
	if version > layout.legacyLast {
		return fmt.Errorf("missing LJC ledger at official version %d: bootstrap from the fork baseline first", version)
	}
	entries, err := os.ReadDir(layout.officialPath)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasSuffix(name, ".up.sql") {
			continue
		}
		prefix, _, ok := strings.Cut(name, "_")
		if !ok {
			continue
		}
		n, e := strconv.Atoi(prefix)
		if e != nil {
			return e
		}
		if n > layout.base {
			return fmt.Errorf("legacy bridge requires frozen upstream migration baseline %d before selecting newer upstream SQL", layout.base)
		}
	}
	applied := version - layout.base
	for _, group := range layout.groups[:applied] {
		for _, table := range group {
			ok, e := ljcTableExists(ctx, tx, layout.sqlite, table)
			if e != nil {
				return e
			}
			if !ok {
				return fmt.Errorf("legacy LJC migration %d lacks required table %s; no ledger changes applied", version, table)
			}
		}
	}
	if !layout.sqlite {
		var n int
		if err := tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM information_schema.columns WHERE table_schema=current_schema() AND table_name='mcp_tool_approvals' AND column_name='enabled'").Scan(&n); err != nil {
			return err
		}
		if n != 0 {
			return fmt.Errorf("official MCP migration marker exists: refusing ambiguous legacy LJC bridge")
		}
		if applied >= 2 {
			if err := tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM information_schema.columns WHERE table_schema=current_schema() AND table_name='task_pending_ops' AND column_name='available_at'").Scan(&n); err != nil {
				return err
			}
			if n != 1 {
				return fmt.Errorf("legacy enrichment migration lacks available_at")
			}
		}
	}
	if _, err = tx.ExecContext(ctx, "CREATE TABLE "+ljcMigrationTable+" (version BIGINT PRIMARY KEY, dirty BOOLEAN NOT NULL)"); err != nil {
		return err
	}
	insert := "INSERT INTO " + ljcMigrationTable + " (version,dirty) VALUES ($1,false)"
	update := "UPDATE schema_migrations SET version=$1 WHERE version=$2 AND dirty=false"
	if layout.sqlite {
		insert = "INSERT INTO " + ljcMigrationTable + " (version,dirty) VALUES (?,false)"
		update = "UPDATE schema_migrations SET version=? WHERE version=? AND dirty=false"
	}
	if _, err = tx.ExecContext(ctx, insert, applied); err != nil {
		return err
	}
	result, err := tx.ExecContext(ctx, update, layout.base, version)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return fmt.Errorf("legacy migration state changed concurrently")
	}
	return tx.Commit()
}

func runLJCMigrations(dsn string, opts MigrationOptions) error {
	layout := ljcLayout(dsn, opts)
	if _, err := os.Stat(layout.path); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}
	db, err := ljcOpenDB(dsn, opts)
	if err != nil {
		return err
	}
	var driver migratedb.Driver
	if layout.sqlite {
		driver, err = sqlite3migrate.WithInstance(db, &sqlite3migrate.Config{MigrationsTable: ljcMigrationTable})
	} else {
		driver, err = pgmigrate.WithInstance(db, &pgmigrate.Config{MigrationsTable: ljcMigrationTable})
	}
	if err != nil {
		db.Close()
		return err
	}
	abs, err := filepath.Abs(layout.path)
	if err != nil {
		driver.Close()
		return err
	}
	name := "postgres"
	if layout.sqlite {
		name = "sqlite3"
	}
	m, err := migrate.NewWithDatabaseInstance("file://"+filepath.ToSlash(abs), name, driver)
	if err != nil {
		driver.Close()
		return err
	}
	defer m.Close()
	if err = m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("LJC migrations: %w", err)
	}
	return nil
}
