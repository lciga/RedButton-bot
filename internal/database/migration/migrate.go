package migration

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

//go:embed migrations/*.up.sql
var migrationFiles embed.FS

const migrationLockID int64 = 821145771

type migration struct {
	version uint64
	name    string
	sql     string
}

// Функция применения невыполненных миграций в порядке версий.
// Каждая миграция и запись о ней сохраняются в одной транзакции.
func Apply(ctx context.Context, db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("apply migrations: nil database")
	}

	db = db.WithContext(ctx)
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version BIGINT PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`).Error; err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	migrations, err := load()
	if err != nil {
		return err
	}

	for _, item := range migrations {
		if err := db.Transaction(func(tx *gorm.DB) error {
			// Блокировка одновременного выполнения миграций несколькими экземплярами приложения.
			if err := tx.Exec("SELECT pg_advisory_xact_lock(?)", migrationLockID).Error; err != nil {
				return err
			}

			var applied int64
			if err := tx.Raw(
				"SELECT COUNT(*) FROM schema_migrations WHERE version = ?",
				item.version,
			).Scan(&applied).Error; err != nil {
				return err
			}
			if applied != 0 {
				return nil
			}

			for _, statement := range splitStatements(item.sql) {
				if err := tx.Exec(statement).Error; err != nil {
					return err
				}
			}
			return tx.Exec(
				"INSERT INTO schema_migrations (version, name) VALUES (?, ?)",
				item.version,
				item.name,
			).Error
		}); err != nil {
			return fmt.Errorf("apply migration %s: %w", item.name, err)
		}
	}

	return nil
}

func load() ([]migration, error) {
	names, err := fs.Glob(migrationFiles, "migrations/*.up.sql")
	if err != nil {
		return nil, fmt.Errorf("list migrations: %w", err)
	}

	items := make([]migration, 0, len(names))
	versions := make(map[uint64]string, len(names))
	for _, path := range names {
		name := strings.TrimPrefix(path, "migrations/")
		prefix, _, ok := strings.Cut(name, "_")
		if !ok {
			return nil, fmt.Errorf("invalid migration filename %q", name)
		}
		version, err := strconv.ParseUint(prefix, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid migration version in %q: %w", name, err)
		}
		if previous, exists := versions[version]; exists {
			return nil, fmt.Errorf("duplicate migration version %d: %q and %q", version, previous, name)
		}
		versions[version] = name

		contents, err := migrationFiles.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read migration %q: %w", name, err)
		}
		items = append(items, migration{version: version, name: name, sql: string(contents)})
	}

	sort.Slice(items, func(i, j int) bool { return items[i].version < items[j].version })
	return items, nil
}

// Функция разделения SQL-файла на отдельные выражения.
// Миграции проекта содержат только DDL без процедурных блоков.
func splitStatements(script string) []string {
	parts := strings.Split(script, ";")
	statements := make([]string, 0, len(parts))
	for _, part := range parts {
		if statement := strings.TrimSpace(part); statement != "" {
			statements = append(statements, statement)
		}
	}
	return statements
}
