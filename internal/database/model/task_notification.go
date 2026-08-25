package model

import (
	"time"

	"github.com/google/uuid"
)

// Структура таблицы отправленных уведомлений о тасках
type TaskNotification struct {
	UserID uuid.UUID `gorm:"type:uuid;primaryKey"` // Идентификатор пользователя
	TaskID uuid.UUID `gorm:"type:uuid;primaryKey"` // Идентификатор таска
	SentAt time.Time `gorm:"not null"`             // Время отправки уведомления
}
