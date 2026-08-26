package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	applicationbot "RedButton-bot/internal/bot"
	"RedButton-bot/internal/config"
	"RedButton-bot/internal/database"
	applicationlogger "RedButton-bot/internal/logger"
	postgresrepository "RedButton-bot/internal/repository/postgres"
	"RedButton-bot/internal/service"
	"RedButton-bot/internal/taskconfig"
)

func main() {
	logger := applicationlogger.New(os.Stdout, os.Getenv("LOG_LEVEL"))
	slog.SetDefault(logger)

	ctx, cancel := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer cancel()

	if err := run(ctx, logger); err != nil {
		if !applicationlogger.IsLogged(err) {
			logger.Error("Application failed", "error", err)
		}
		os.Exit(1)
	}
}

func run(ctx context.Context, logger *slog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}
	tasks, err := taskconfig.Load(cfg.TasksDirectory, cfg.TimeZone)
	if err != nil {
		return fmt.Errorf("load task configuration: %w", err)
	}

	db, err := database.Open(ctx, cfg.DatabaseDSN, logger)
	if err != nil {
		return err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("get PostgreSQL connection: %w", err)
	}
	defer func() {
		if err := sqlDB.Close(); err != nil {
			logger.Error("Failed to close PostgreSQL connection", "error", err)
		}
	}()

	if err := database.Migrate(ctx, db); err != nil {
		return err
	}

	store := postgresrepository.New(db)
	services := service.New(store.Repositories(), store, nil, cfg.AdminTelegramIDs)
	if err := services.Tasks.Sync(
		ctx,
		tasks,
		cfg.TaskExpire,
		cfg.BotStartDate,
		cfg.BotEndDate,
	); err != nil {
		return fmt.Errorf("synchronize tasks: %w", err)
	}
	telegramBot, err := applicationbot.New(
		cfg.TelegramBotToken,
		services,
		logger,
		cfg.NotificationInterval,
		cfg.TelegramInitTimeout,
		cfg.AdminTelegramIDs,
		cfg.BotStartDate,
		cfg.BotEndDate,
	)
	if err != nil {
		return err
	}

	logger.Info(
		"Application started",
		"tasks", len(tasks),
		"admins", len(cfg.AdminTelegramIDs),
		"starts_at", cfg.BotStartDate,
		"ends_at", cfg.BotEndDate,
	)
	telegramBot.Start(ctx)
	return nil
}
