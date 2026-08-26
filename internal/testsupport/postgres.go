package testsupport

import (
	"context"
	"io"
	"log/slog"
	"os"
	"strings"
	"testing"

	"RedButton-bot/internal/database"
	"gorm.io/gorm"
)

// PostgreSQL opens the explicitly configured test database and clears application tables.
// The database name must end in _test to prevent accidental damage to non-test data.
func PostgreSQL(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("TEST_DATABASE_DSN is not configured")
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	db, err := database.Open(context.Background(), dsn, logger)
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	var name string
	if err := db.Raw("SELECT current_database()").Scan(&name).Error; err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(name, "_test") {
		t.Fatalf("refusing to clear database %q: test database name must end in _test", name)
	}
	if err := database.Migrate(context.Background(), db); err != nil {
		t.Fatal(err)
	}
	if err := db.Exec("TRUNCATE task_notifications, submissions, ratings, task_files, tasks, users CASCADE").Error; err != nil {
		t.Fatal(err)
	}
	return db
}
