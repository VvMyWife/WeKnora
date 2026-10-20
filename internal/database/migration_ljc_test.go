package database

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLJCMigrationsFreshAndLegacySQLite(t *testing.T) {
	root := sqliteRepoRoot(t)
	legacy := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(legacy, "migrations/sqlite"), 0755))
	files, err := filepath.Glob(filepath.Join(root, "migrations/ljc/bootstrap/sqlite/*.sql"))
	require.NoError(t, err)
	for _, f := range files {
		b, e := os.ReadFile(f)
		require.NoError(t, e)
		require.NoError(t, os.WriteFile(filepath.Join(legacy, "migrations/sqlite", filepath.Base(f)), b, 0644))
	}
	for _, mapping := range [][2]string{
		{"000001_archive_mapping", "000013_archive_mapping"},
		{"000002_external_chunks", "000014_external_chunks"},
		{"000003_archival_citations", "000015_archival_citations"},
	} {
		for _, direction := range []string{".up.sql", ".down.sql"} {
			b, e := os.ReadFile(filepath.Join(root, "migrations/ljc/sqlite", mapping[0]+direction))
			require.NoError(t, e)
			require.NoError(t, os.WriteFile(filepath.Join(legacy, "migrations/sqlite", mapping[1]+direction), b, 0644))
		}
	}
	for _, scenario := range []string{"fresh", "legacy", "missing-table", "dirty"} {
		t.Run(scenario, func(t *testing.T) {
			dbPath := filepath.Join(t.TempDir(), "test.db")
			opts := MigrationOptions{SQLiteDBPath: dbPath}
			if scenario != "fresh" {
				chdirAndRestore(t, legacy)
				require.NoError(t, RunMigrationsWithOptions("sqlite3://unused", opts))
				db := openSQLiteDB(t, dbPath)
				version, dirty := sqliteMigrationState(t, db)
				require.Equal(t, 15, version)
				require.False(t, dirty)
				_, err := db.Exec("INSERT INTO archive_mapping_links(knowledge_id,tenant_id,version_id) VALUES ('sentinel',1,'original')")
				require.NoError(t, err)
				if scenario == "missing-table" {
					_, err = db.Exec("DROP TABLE external_chunk_records")
					require.NoError(t, err)
				}
				if scenario == "dirty" {
					_, err = db.Exec("UPDATE schema_migrations SET dirty=true")
					require.NoError(t, err)
				}
			}
			chdirAndRestore(t, root)
			err := BootstrapLJCLegacy("sqlite3://unused", opts)
			db := openSQLiteDB(t, dbPath)
			if scenario == "missing-table" || scenario == "dirty" {
				require.Error(t, err)
				version, _ := sqliteMigrationState(t, db)
				require.Equal(t, 15, version)
				require.False(t, sqliteTableExists(t, db, ljcMigrationTable))
				return
			}
			require.NoError(t, err)
			require.NoError(t, RunMigrationsWithOptions("sqlite3://unused", opts))
			version, dirty := sqliteMigrationState(t, db)
			require.Equal(t, 45, version)
			require.False(t, dirty)
			var customVersion int
			var customDirty bool
			require.NoError(t, db.QueryRow("SELECT version,dirty FROM "+ljcMigrationTable).Scan(&customVersion, &customDirty))
			require.Equal(t, 3, customVersion)
			require.False(t, customDirty)
			if scenario == "legacy" {
				var value string
				require.NoError(t, db.QueryRow("SELECT version_id FROM archive_mapping_links WHERE knowledge_id='sentinel'").Scan(&value))
				require.Equal(t, "original", value)
			}
		})
	}
}

func TestLJCMigrationPostgresRestore(t *testing.T) {
	dsn := os.Getenv("LJC_MIGRATION_TEST_DSN")
	if dsn == "" {
		t.Skip("requires dedicated restored PostgreSQL; never use the production DSN")
	}
	root := sqliteRepoRoot(t)
	chdirAndRestore(t, root)
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	defer db.Close()
	var database string
	require.NoError(t, db.QueryRow("SELECT current_database()").Scan(&database))
	require.Equal(t, "ljc_migration_test", database, "integration test only accepts its dedicated database name")
	var beforeVersion int
	var beforeDirty bool
	require.NoError(t, db.QueryRow("SELECT version,dirty FROM schema_migrations").Scan(&beforeVersion, &beforeDirty))
	require.Equal(t, 94, beforeVersion)
	require.False(t, beforeDirty)
	counts := map[string]int64{}
	for _, table := range []string{"knowledges", "chunks", "wiki_pages", "archive_mapping_versions", "message_citation_evidence", "enrichment_recoveries"} {
		var n int64
		require.NoError(t, db.QueryRow("SELECT COUNT(*) FROM "+table).Scan(&n))
		counts[table] = n
	}
	require.NoError(t, BootstrapLJCLegacy(dsn, MigrationOptions{}))
	require.NoError(t, RunMigrationsWithOptions(dsn, MigrationOptions{}))
	require.NoError(t, RunMigrationsWithOptions(dsn, MigrationOptions{}))
	var version, custom int
	var dirty bool
	require.NoError(t, db.QueryRow("SELECT version,dirty FROM schema_migrations").Scan(&version, &dirty))
	require.Equal(t, 128, version)
	require.False(t, dirty)
	require.NoError(t, db.QueryRow("SELECT version,dirty FROM "+ljcMigrationTable).Scan(&custom, &dirty))
	require.Equal(t, 4, custom)
	require.False(t, dirty)
	for table, want := range counts {
		var got int64
		require.NoError(t, db.QueryRow("SELECT COUNT(*) FROM "+table).Scan(&got))
		require.Equal(t, want, got, table)
	}
}
