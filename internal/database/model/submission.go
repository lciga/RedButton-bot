package model

import (
	"time"

	"github.com/google/uuid"
)

// Структура таблицы с попытками сдачи
type Submission struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`                           // Уникальный идентификатор попытки
	UserID        uuid.UUID `gorm:"type:uuid;not null;index"`                                                 // Идентификатор пользователя
	User          User      `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"` // Связь с пользователем
	TaskID        uuid.UUID `gorm:"type:uuid;not null;index"`                                                 // Идентификатор таска
	Task          Task      `gorm:"foreignKey:TaskID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"` // Связь с таском
	IsCorrect     bool      `gorm:"not null;default:false;index"`                                             // Флаг корректности попытки
	PointsAwarded int       `gorm:"not null;default:0"`                                                       // Количество начисленных очков
	SubmittedAt   time.Time `gorm:"not null;index"`                                                           // Время сдачи
}
