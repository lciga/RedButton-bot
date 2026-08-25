package dto

import "github.com/google/uuid"

// Структура данных уведомления о новом таске
type TaskNotification struct {
	UserID         uuid.UUID // Идентификатор пользователя
	TelegramUserID int64     // Идентификатор пользователя в Telegram
	Task           Task      // Данные нового таска
}
