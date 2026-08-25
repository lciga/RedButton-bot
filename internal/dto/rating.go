package dto

import (
	"time"

	"github.com/google/uuid"
)

// Структура данных позиции в рейтинге
type Rating struct {
	Position         int        // Позиция пользователя в рейтинге
	UserID           uuid.UUID  // Идентификатор пользователя
	TelegramUserID   int64      // Идентификатор пользователя в Telegram
	TelegramUsername *string    // Имя пользователя в Telegram
	FirstName        *string    // Имя пользователя
	LastName         *string    // Фамилия пользователя
	TotalPoints      int        // Общее количество заработанных очков
	SolvedTasksCount int        // Количество решённых тасок
	LastSolvedAt     *time.Time // Время последнего решения
}

// Структура данных страницы рейтинга
type RatingPage struct {
	Items      []Rating // Позиции пользователей
	Page       int      // Текущая страница
	TotalPages int      // Общее количество страниц
	TotalItems int64    // Общее количество участников
}
