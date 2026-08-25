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
	postgresrepository "RedButton-bot/internal/repository/postgres"
	"RedButton-bot/internal/service"
	"RedButton-bot/internal/taskconfig"
)

func main() {
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer cancel()

	if err := run(ctx); err != nil {
		slog.Error("ошибка работы приложения", "error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("загрузить конфиг: %w", err)
	}
	tasks, err := taskconfig.Load(cfg.TasksDirectory)
	if err != nil {
		return fmt.Errorf("загрузить таски: %w", err)
	}

	db, err := database.Open(ctx, cfg.DatabaseDSN)
	if err != nil {
		return err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("получить соединение PostgreSQL: %w", err)
	}
	defer func() {
		if err := sqlDB.Close(); err != nil {
			slog.Error("не удалось закрыть соединение PostgreSQL", "error", err)
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
		return fmt.Errorf("синхронизировать таски: %w", err)
	}
	telegramBot, err := applicationbot.New(
		cfg.TelegramBotToken,
		services,
		slog.Default(),
		cfg.NotificationInterval,
		cfg.AdminTelegramIDs,
		cfg.BotStartDate,
		cfg.BotEndDate,
	)
	if err != nil {
		return err
	}

	slog.Info("Telegram-бот запущен")
	telegramBot.Start(ctx)
	return nil
}
