package config

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

var utcPlusFive = time.FixedZone("UTC+5", 5*60*60)
var utcOffsetPattern = regexp.MustCompile(`^UTC([+-])(\d{1,2})(?::?(\d{2}))?$`)

// Структура конфига бота
type Config struct {
	TelegramBotToken       string             // Токен бота
	TelegramInitTimeout    time.Duration      // Таймаут проверки подключения к Telegram
	FlagSubmissionInterval time.Duration      // Минимальный интервал между попытками сдачи флага
	DatabaseDSN            string             // Строка подключения к PostgreSQL
	TasksDirectory         string             // Путь к директории с YAML-файлами тасков
	TimeZone               *time.Location     // Часовой пояс времени работы бота
	BotStartDate           time.Time          // Дата и время начала работы бота
	BotEndDate             time.Time          // Дата и время окончания работы бота
	TaskExpire             time.Duration      // Время жизни тасок
	NotificationInterval   time.Duration      // Интервал проверки новых тасков
	AdminTelegramIDs       map[int64]struct{} // Идентификаторы администраторов в Telegram
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
	cfg.TelegramInitTimeout = 30 * time.Second
	if value := os.Getenv("TELEGRAM_INIT_TIMEOUT"); value != "" {
		cfg.TelegramInitTimeout, err = time.ParseDuration(value)
		if err != nil {
			return nil, fmt.Errorf("parse TELEGRAM_INIT_TIMEOUT: %w", err)
		}
	}
	if cfg.TelegramInitTimeout <= 0 {
		return nil, fmt.Errorf("TELEGRAM_INIT_TIMEOUT must be greater than zero")
	}
	cfg.FlagSubmissionInterval = 5 * time.Second
	if value := os.Getenv("FLAG_SUBMISSION_INTERVAL"); value != "" {
		cfg.FlagSubmissionInterval, err = time.ParseDuration(value)
		if err != nil {
			return nil, fmt.Errorf("parse FLAG_SUBMISSION_INTERVAL: %w", err)
		}
	}
	if cfg.FlagSubmissionInterval <= 0 {
		return nil, fmt.Errorf("FLAG_SUBMISSION_INTERVAL must be greater than zero")
	}
	cfg.DatabaseDSN = os.Getenv("DATABASE_DSN")
	cfg.TasksDirectory = os.Getenv("TASKS_DIRECTORY")
	if cfg.TasksDirectory == "" {
		cfg.TasksDirectory = "tasks"
	}
	cfg.TimeZone = utcPlusFive
	if value := os.Getenv("TASK_TIMEZONE"); value != "" {
		cfg.TimeZone, err = loadTimeZone(value)
		if err != nil {
			return nil, fmt.Errorf("load TASK_TIMEZONE: %w", err)
		}
	}
	cfg.BotStartDate, err = time.ParseInLocation("2006-01-02T15:04:05", os.Getenv("TASK_START_DATE"), cfg.TimeZone)
	if err != nil {
		return nil, fmt.Errorf("parse TASK_START_DATE: %w", err)
	}
	cfg.BotEndDate, err = time.ParseInLocation("2006-01-02T15:04:05", os.Getenv("TASK_END_DATE"), cfg.TimeZone)
	if err != nil {
		return nil, fmt.Errorf("parse TASK_END_DATE: %w", err)
	}
	if !cfg.BotEndDate.After(cfg.BotStartDate) {
		return nil, fmt.Errorf("bot end time must be later than start time")
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

func loadTimeZone(value string) (*time.Location, error) {
	if value == "UTC" {
		return time.UTC, nil
	}
	match := utcOffsetPattern.FindStringSubmatch(value)
	if match == nil {
		return time.LoadLocation(value)
	}

	hours, _ := strconv.Atoi(match[2])
	minutes := 0
	if match[3] != "" {
		minutes, _ = strconv.Atoi(match[3])
	}
	if hours > 14 || minutes >= 60 || hours == 14 && minutes != 0 {
		return nil, fmt.Errorf("invalid UTC offset %q", value)
	}
	offset := (hours*60 + minutes) * 60
	if match[1] == "-" {
		offset = -offset
	}
	return time.FixedZone(value, offset), nil
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
			return nil, fmt.Errorf("invalid administrator Telegram ID %q", item)
		}
		result[telegramID] = struct{}{}
	}

	return result, nil
}
