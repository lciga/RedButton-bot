package model

import (
	"time"

	"github.com/google/uuid"
)

// Структура таблицы с файлами тасков
type TaskFile struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`                           // Уникальный идентификатор файла
	TaskID         uuid.UUID `gorm:"type:uuid;not null;index"`                                                 // Идентификатор таска
	Task           Task      `gorm:"foreignKey:TaskID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"` // Связь с таском
	StoragePath    string    `gorm:"type:text;not null"`                                                       // Путь хранения файла
	FileName       string    `gorm:"size:255;not null"`                                                        // Имя файла
	TelegramFileID *string   `gorm:"type:text"`                                                                // Идентификатор файла в Telegram
	MIMEType       *string   `gorm:"size:255"`                                                                 // MIME-тип файла
	FileSize       int64     `gorm:"not null"`                                                                 // Размер файла в байтах
	CreatedAt      time.Time `gorm:"not null"`                                                                 // Время создания записи
}
