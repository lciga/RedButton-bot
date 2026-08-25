package dto

import (
	"time"

	"github.com/google/uuid"
)

// Структура данных таска
type Task struct {
	ID            uuid.UUID  // Уникальный идентификатор таска
	Slug          string     // Идентификатор таска
	Title         string     // Название таска
	Description   string     // Описание таска
	CurrentPoints int        // Текущая стоимость таска
	StartsAt      time.Time  // Время открытия таска
	EndsAt        *time.Time // Время закрытия таска
	Files         []TaskFile // Прикреплённые файлы
}

// Структура данных файла таска
type TaskFile struct {
	ID             uuid.UUID // Уникальный идентификатор файла
	FileName       string    // Имя файла
	StoragePath    string    // Путь хранения файла
	TelegramFileID *string   // Идентификатор файла в Telegram
	MIMEType       *string   // MIME-тип файла
	FileSize       int64     // Размер файла в байтах
}

// Структура данных таска из YAML-файла
type TaskDefinition struct {
	Slug          string              // Идентификатор таска
	Title         string              // Название таска
	Description   string              // Описание таска
	Flag          string              // Правильный флаг
	MaximumPoints int                 // Максимальная стоимость таска
	MinimumPoints int                 // Минимальная стоимость таска
	Decay         int                 // Количество решений для полного распада
	StartsAt      time.Time           // Время открытия таска
	File          *TaskFileDefinition // Прикреплённый файл
}

// Структура данных файла из YAML-файла
type TaskFileDefinition struct {
	StoragePath string  // Полный путь к файлу
	FileName    string  // Имя файла
	MIMEType    *string // MIME-тип файла
	FileSize    int64   // Размер файла в байтах
}
