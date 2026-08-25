package database

import (
	"context"
	"fmt"
	"log/slog"

	"RedButton-bot/internal/database/migration"
	applicationlogger "RedButton-bot/internal/logger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Функция подключения к PostgreSQL и проверки соединения.
// Возвращает экземпляр GORM и ошибку.
func Open(ctx context.Context, dsn string, logger *slog.Logger) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: applicationlogger.GORM(logger),
	})
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get postgres connection: %w", err)
	}
	if err := sqlDB.PingContext(ctx); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return db, nil
}

// Функция применения встроенных SQL-миграций.
// Возвращает ошибку выполнения миграций.
func Migrate(ctx context.Context, db *gorm.DB) error {
	if err := migration.Apply(ctx, db); err != nil {
		return fmt.Errorf("migrate database: %w", err)
	}
	return nil
}
