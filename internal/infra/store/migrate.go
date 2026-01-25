package store

import (
	"database/sql"
	"embed"
	"fmt"
	"sort"
	"strings"
)

//go:embed migrations/*.sql
var migrations embed.FS

// Migrate runs all embedded SQL migrations in order.
func Migrate(db *sql.DB) error {
	entries, err := migrations.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	// Sort entries by name to ensure order
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		content, err := migrations.ReadFile("migrations/" + entry.Name())
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", entry.Name(), err)
		}

		if _, err := db.Exec(string(content)); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", entry.Name(), err)
		}
	}

	// Run programmatic migrations for columns that can't be added idempotently in SQL
	if err := migrateGitColumns(db); err != nil {
		return fmt.Errorf("failed to migrate git columns: %w", err)
	}

	return nil
}

// migrateGitColumns adds the branch and base_branch columns if they don't exist.
func migrateGitColumns(db *sql.DB) error {
	// Check if columns exist
	rows, err := db.Query("PRAGMA table_info(agents)")
	if err != nil {
		return err
	}
	defer rows.Close()

	hasBranch := false
	hasBaseBranch := false

	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dfltValue interface{}
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk); err != nil {
			continue
		}
		if strings.EqualFold(name, "branch") {
			hasBranch = true
		}
		if strings.EqualFold(name, "base_branch") {
			hasBaseBranch = true
		}
	}

	if !hasBranch {
		if _, err := db.Exec("ALTER TABLE agents ADD COLUMN branch TEXT DEFAULT ''"); err != nil {
			return err
		}
	}

	if !hasBaseBranch {
		if _, err := db.Exec("ALTER TABLE agents ADD COLUMN base_branch TEXT DEFAULT ''"); err != nil {
			return err
		}
	}

	return nil
}
