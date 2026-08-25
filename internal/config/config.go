package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

var utcPlusFive = time.FixedZone("UTC+5", 5*60*60)

// Структура конфига бота
type Config struct {
	TelegramBotToken     string             // Токен бота
	DatabaseDSN          string             // Строка подключения к PostgreSQL
	TasksDirectory       string             // Путь к директории с YAML-файлами тасков
	BotStartDate         time.Time          // Дата и время начала работы бота
	BotEndDate           time.Time          // Дата и время окончания работы бота
	TaskExpire           time.Duration      // Время жизни тасок
	NotificationInterval time.Duration      // Интервал проверки новых тасков
	AdminTelegramIDs     map[int64]struct{} // Идентификаторы администраторов в Telegram
}

// Функция загрузки конфига.
// Возвращает указатель на экземпляр Config и ошибку.
func Load() (*Config, error) {
	err := godotenv.Load()
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	cfg := &Config{}

	cfg.TelegramBotToken = os.Getenv("TELEGRAM_BOT_TOKEN")
	cfg.DatabaseDSN = os.Getenv("DATABASE_DSN")
	cfg.TasksDirectory = os.Getenv("TASKS_DIRECTORY")
	if cfg.TasksDirectory == "" {
		cfg.TasksDirectory = "tasks"
	}
	cfg.BotStartDate, err = time.ParseInLocation("2006-01-02T15:04:05", os.Getenv("TASK_START_DATE"), utcPlusFive)
	if err != nil {
		return nil, err
	}
	cfg.BotEndDate, err = time.ParseInLocation("2006-01-02T15:04:05", os.Getenv("TASK_END_DATE"), utcPlusFive)
	if err != nil {
		return nil, err
	}
	if !cfg.BotEndDate.After(cfg.BotStartDate) {
		return nil, fmt.Errorf("время окончания работы бота должно быть позже времени начала")
	}
	cfg.TaskExpire, err = time.ParseDuration(os.Getenv("TASK_EXPIRE"))
	if err != nil {
		return nil, err
	}
	cfg.NotificationInterval = 15 * time.Second
	if value := os.Getenv("NOTIFICATION_INTERVAL"); value != "" {
		cfg.NotificationInterval, err = time.ParseDuration(value)
		if err != nil {
			return nil, err
		}
	}
	cfg.AdminTelegramIDs, err = parseTelegramIDs(os.Getenv("ADMIN_TELEGRAM_IDS"))
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func parseTelegramIDs(value string) (map[int64]struct{}, error) {
	result := make(map[int64]struct{})
	for _, item := range strings.Split(value, ",") {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}

		telegramID, err := strconv.ParseInt(item, 10, 64)
		if err != nil || telegramID <= 0 {
			return nil, fmt.Errorf("некорректный Telegram ID администратора %q", item)
		}
		result[telegramID] = struct{}{}
	}

	return result, nil
}
