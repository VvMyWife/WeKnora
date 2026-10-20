package database

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// BootstrapLJCLegacy explicitly splits the historical custom ledger using a
// pinned SQL baseline. It runs no application jobs and applies no new schema.
func BootstrapLJCLegacy(dsn string, opts MigrationOptions) error {
	root := filepath.Join("migrations", "ljc", "bootstrap")
	raw, err := os.ReadFile(filepath.Join(root, "manifest.json"))
	if err != nil {
		return fmt.Errorf("read pinned bootstrap manifest: %w", err)
	}
	var manifest map[string]string
	if err := json.Unmarshal(raw, &manifest); err != nil {
		return err
	}
	if len(manifest) == 0 {
		return fmt.Errorf("empty bootstrap manifest")
	}
	for name, want := range manifest {
		clean := filepath.Clean(name)
		if filepath.IsAbs(clean) || clean == ".." || len(clean) >= 3 && clean[:3] == ".."+string(filepath.Separator) {
			return fmt.Errorf("invalid bootstrap path")
		}
		contents, err := os.ReadFile(filepath.Join(root, clean))
		if err != nil {
			return err
		}
		sum := sha256.Sum256(contents)
		if hex.EncodeToString(sum[:]) != want {
			return fmt.Errorf("bootstrap baseline changed: %s", name)
		}
	}
	layout := ljcLayout(dsn, opts)
	dialect := "postgres"
	if layout.sqlite {
		dialect = "sqlite"
	}
	layout.officialPath = filepath.Join(root, dialect)
	return prepareLJCLegacyBridgeWithLayout(dsn, opts, layout)
}

// LJCMigrationStatus reads both ledgers without running any migration.
func LJCMigrationStatus(dsn string, opts MigrationOptions) (map[string]interface{}, error) {
	db, err := ljcOpenDB(dsn, opts)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	tx, err := db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	layout := ljcLayout(dsn, opts)
	engine := "postgres"
	if layout.sqlite {
		engine = "sqlite"
	}
	result := map[string]interface{}{"engine": engine}
	for key, table := range map[string]string{"official": "schema_migrations", "ljc": ljcMigrationTable} {
		exists, err := ljcTableExists(ctx, tx, layout.sqlite, table)
		if err != nil {
			return nil, err
		}
		entry := map[string]interface{}{"exists": exists, "version": nil, "dirty": false}
		if exists {
			var version int64
			var dirty bool
			err := tx.QueryRowContext(ctx, "SELECT version,dirty FROM "+table).Scan(&version, &dirty)
			if err != nil && err != sql.ErrNoRows {
				return nil, err
			}
			if err == nil {
				entry["version"] = version
				entry["dirty"] = dirty
			}
		}
		result[key] = entry
	}
	return result, nil
}
