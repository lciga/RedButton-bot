package model

import (
	"time"

	"github.com/google/uuid"
)

// Структура таблицы пользователей
type User struct {
	ID               uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"` // Уникальный идентификатор пользователя
	TelegramUserID   int64      `gorm:"not null;uniqueIndex"`                           // Идентификатор пользователя в Telegram
	TelegramUsername *string    `gorm:"size:255"`                                       // Имя пользователя в Telegram
	FirstName        *string    `gorm:"size:255"`                                       // Имя пользователя
	LastName         *string    `gorm:"size:255"`                                       // Фамилия пользователя
	IsActive         bool       `gorm:"not null;default:true"`                          // Флаг активности пользователя
	CreatedAt        time.Time  `gorm:"not null"`                                       // Время создания пользователя
	UpdatedAt        time.Time  `gorm:"not null"`                                       // Время обновления пользователя
	LastSeenAt       *time.Time // Время последней активности пользователя
}
