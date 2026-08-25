package dto

import (
	"time"

	"github.com/google/uuid"
)

// Структура данных авторизации пользователя
type AuthenticateUser struct {
	TelegramUserID   int64   // Идентификатор пользователя в Telegram
	TelegramUsername *string // Имя пользователя в Telegram
	FirstName        *string // Имя пользователя
	LastName         *string // Фамилия пользователя
}

// Структура данных пользователя
type User struct {
	ID               uuid.UUID  // Уникальный идентификатор пользователя
	TelegramUserID   int64      // Идентификатор пользователя в Telegram
	TelegramUsername *string    // Имя пользователя в Telegram
	FirstName        *string    // Имя пользователя
	LastName         *string    // Фамилия пользователя
	IsActive         bool       // Флаг активности пользователя
	LastSeenAt       *time.Time // Время последней активности пользователя
}
