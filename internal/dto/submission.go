package dto

import "github.com/google/uuid"

// Структура данных попытки решения
type SubmitTask struct {
	TelegramUserID int64     // Идентификатор пользователя в Telegram
	TaskID         uuid.UUID // Идентификатор таска
	Flag           string    // Флаг пользователя
}

// Структура результата попытки решения
type SubmissionResult struct {
	TaskID        uuid.UUID // Идентификатор таска
	Correct       bool      // Флаг корректности решения
	AlreadySolved bool      // Флаг повторного решения
	Ignored       bool      // Флаг попытки без сохранения результата
	PointsAwarded int       // Количество начисленных очков
	TotalPoints   int       // Общее количество очков пользователя
	CurrentPoints int       // Текущая стоимость таска после решения
}
