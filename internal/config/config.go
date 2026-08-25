package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

var utcPlusFive = time.FixedZone("UTC+5", 5*60*60)

// Структура конфига бота
type Config struct {
	TelegramBotToken string        // Токен бота
	TaskStartDate    time.Time     // Дата и время запуска решения тасок
	TaskEndDate      time.Time     // Дата и время окончания решения тасок
	TaskExpire       time.Duration // Время жизни тасок
	MaxPointsValue   int           // Максимальное число очков
	MinPointsValue   int           // Минимальное число очков
	Decay            int           // Количество решений, после которого начинается распад
}

// Функция загрузки конфига.
// Возвращает указатель на экземпляр Config и ошибку.
func Load() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}

	cfg := &Config{}

	cfg.TelegramBotToken = os.Getenv("TELEGRAM_BOT_TOKEN")
	cfg.TaskStartDate, err = time.ParseInLocation("2006-01-02T15:04:05", os.Getenv("TASK_START_DATE"), utcPlusFive)
	if err != nil {
		return nil, err
	}
	cfg.TaskEndDate, err = time.ParseInLocation("2006-01-02T15:04:05", os.Getenv("TASK_END_DATE"), utcPlusFive)
	if err != nil {
		return nil, err
	}
	cfg.MaxPointsValue, err = strconv.Atoi(os.Getenv("MAX_POINTS_VALUE"))
	if err != nil {
		return nil, err
	}
	cfg.MinPointsValue, err = strconv.Atoi(os.Getenv("MIN_POINTS_VALUE"))
	if err != nil {
		return nil, err
	}
	cfg.Decay, err = strconv.Atoi(os.Getenv("DECAY"))
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
