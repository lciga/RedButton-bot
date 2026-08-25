package model

import (
	"time"

	"github.com/google/uuid"
)

// Структура таблицы с тасками
type Task struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"` // Уникальный идентификатор таска
	Slug          string     `gorm:"size:255;not null;uniqueIndex"`                  // Идентификатор для синхронизации таска из YAML
	Title         string     `gorm:"size:255;not null"`                              // Название таска
	Description   string     `gorm:"type:text;not null"`                             // Описание таска
	FlagHash      string     `gorm:"size:255;not null"`                              // Хэш правильного флага
	InitialPoints int        `gorm:"not null"`                                       // Начальное количество очков
	MinimumPoints int        `gorm:"not null"`                                       // Минимальное количество очков
	CurrentPoints int        `gorm:"not null"`                                       // Текущая стоимость таска
	Decay         int        `gorm:"not null"`                                       // Число решений, после которого начинается распад
	StartsAt      time.Time  `gorm:"not null;index"`                                 // Время открытия таска
	EndsAt        *time.Time // Время закрытия таска
	IsActive      bool       `gorm:"not null;default:true;index"` // Флаг активности таска
	CreatedAt     time.Time  `gorm:"not null"`                    // Время создания таска
	UpdatedAt     time.Time  `gorm:"not null"`                    // Время обновления таска
	Files         []TaskFile `gorm:"foreignKey:TaskID" json:"-"`  // Связь с файлами таска
}
