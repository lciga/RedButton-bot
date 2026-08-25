package model

import (
	"time"

	"github.com/google/uuid"
)

// Структура таблицы рейтинга
type Rating struct {
	UserID           uuid.UUID  `gorm:"type:uuid;primaryKey"`                                                     // Идентификатор пользователя
	User             User       `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"` // Связь с пользователем
	TotalPoints      int        `gorm:"not null;default:0;index"`                                                 // Общее количество заработанных очков
	SolvedTasksCount int        `gorm:"not null;default:0"`                                                       // Количество решённых тасок
	LastSolvedAt     *time.Time // Время последнего решения
	UpdatedAt        time.Time  `gorm:"not null"` // Время обновления рейтинга
}
