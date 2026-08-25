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
		logger.Error("Application stopped with an error", "error", err, "error_type", fmt.Sprintf("%T", err))
		os.Exit(1)
	}
}

func run(ctx context.Context, logger *slog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("загрузить конфиг: %w", err)
	}
	logger.Info(
		"Configuration loaded",
		"tasks_directory", cfg.TasksDirectory,
		"bot_starts_at", cfg.BotStartDate,
		"bot_ends_at", cfg.BotEndDate,
		"telegram_init_timeout", cfg.TelegramInitTimeout,
	)
	tasks, err := taskconfig.Load(cfg.TasksDirectory)
	if err != nil {
		return fmt.Errorf("загрузить таски: %w", err)
	}
	logger.Info("Task configuration loaded", "tasks_count", len(tasks))

	logger.Info("Connecting to PostgreSQL")
	db, err := database.Open(ctx, cfg.DatabaseDSN, logger)
	if err != nil {
		return err
	}
	logger.Info("PostgreSQL connection established")
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("получить соединение PostgreSQL: %w", err)
	}
	defer func() {
		if err := sqlDB.Close(); err != nil {
			logger.Error("Failed to close PostgreSQL connection", "error", err)
		}
	}()

	if err := database.Migrate(ctx, db); err != nil {
		return err
	}
	logger.Info("Database migrations applied")

	store := postgresrepository.New(db)
	services := service.New(store.Repositories(), store, nil, cfg.AdminTelegramIDs)
	if err := services.Tasks.Sync(
		ctx,
		tasks,
		cfg.TaskExpire,
		cfg.BotStartDate,
		cfg.BotEndDate,
	); err != nil {
		return fmt.Errorf("синхронизировать таски: %w", err)
	}
	logger.Info("Tasks synchronized with database", "tasks_count", len(tasks))
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

	logger.Info("Telegram bot started")
	telegramBot.Start(ctx)
	return nil
}
