package database

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLJCBootstrapChecksPinnedBaselineAndReadsStatus(t *testing.T) {
	root := sqliteRepoRoot(t)
	chdirAndRestore(t, root)
	path := filepath.Join(t.TempDir(), "fresh.db")
	opts := MigrationOptions{SQLiteDBPath: path}
	require.NoError(t, BootstrapLJCLegacy("sqlite3://unused", opts))
	status, err := LJCMigrationStatus("sqlite3://unused", opts)
	require.NoError(t, err)
	require.Equal(t, false, status["official"].(map[string]interface{})["exists"])
	require.Equal(t, false, status["ljc"].(map[string]interface{})["exists"])
}

func TestLJCBootstrapRejectsTamperedPinnedBaseline(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "migrations/ljc/bootstrap/postgres")
	require.NoError(t, os.MkdirAll(dir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "000090_baseline.up.sql"), []byte("altered"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(root, "migrations/ljc/bootstrap/manifest.json"), []byte("{\"postgres/000090_baseline.up.sql\":\"invalid\"}"), 0644))
	chdirAndRestore(t, root)
	err := BootstrapLJCLegacy("sqlite3://unused", MigrationOptions{SQLiteDBPath: filepath.Join(t.TempDir(), "test.db")})
	require.ErrorContains(t, err, "bootstrap baseline changed")
}
