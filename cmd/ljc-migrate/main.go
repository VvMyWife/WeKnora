// Command ljc-migrate manages schema only; it never starts workers or schedulers.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/Tencent/WeKnora/internal/database"
)

func main() {
	mode := flag.String("mode", "status", "status, bootstrap-legacy, or up")
	flag.Parse()
	dsn := os.Getenv("LJC_MIGRATION_DSN")
	sqlite := os.Getenv("LJC_SQLITE_PATH")
	if dsn == "" && sqlite == "" {
		fail(fmt.Errorf("set LJC_MIGRATION_DSN or LJC_SQLITE_PATH; never pass credentials as arguments"))
	}
	opts := database.MigrationOptions{SQLiteDBPath: sqlite, AutoRecoverDirty: false}
	if sqlite != "" {
		if dsn != "" {
			fail(fmt.Errorf("set only one database target"))
		}
		dsn = "sqlite3://" + sqlite
	}
	var err error
	switch *mode {
	case "status":
	case "bootstrap-legacy":
		err = database.BootstrapLJCLegacy(dsn, opts)
	case "up":
		err = database.RunMigrationsWithOptions(dsn, opts)
	default:
		err = fmt.Errorf("unsupported mode %q", *mode)
	}
	if err != nil {
		fail(err)
	}
	status, err := database.LJCMigrationStatus(dsn, opts)
	if err != nil {
		fail(err)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(status); err != nil {
		fail(err)
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
