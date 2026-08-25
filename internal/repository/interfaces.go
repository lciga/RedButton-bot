package repository

import (
	"context"
	"errors"
	"time"

	"RedButton-bot/internal/database/model"
	"github.com/google/uuid"
)

var ErrNotFound = errors.New("запись не найдена")

// Интерфейс репозитория пользователей
type UserRepository interface {
	Upsert(ctx context.Context, user *model.User) error
	GetByTelegramID(ctx context.Context, telegramUserID int64) (*model.User, error)
}

// Интерфейс репозитория тасков
type TaskRepository interface {
	Sync(ctx context.Context, tasks []model.Task) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Task, error)
	GetByIDForUpdate(ctx context.Context, id uuid.UUID) (*model.Task, error)
	GetNewestAvailableForUser(ctx context.Context, telegramUserID int64, now time.Time) (*model.Task, error)
	GetNextStartsAt(ctx context.Context, now time.Time) (*time.Time, error)
	ListUpcoming(ctx context.Context, now time.Time, limit int) ([]model.Task, error)
	ListAvailable(ctx context.Context, now time.Time) ([]model.Task, error)
	UpdateCurrentPoints(ctx context.Context, id uuid.UUID, points int) error
}

// Интерфейс репозитория попыток сдачи
type SubmissionRepository interface {
	Create(ctx context.Context, submission *model.Submission) error
	HasCorrect(ctx context.Context, userID, taskID uuid.UUID) (bool, error)
	CountCorrect(ctx context.Context, taskID uuid.UUID) (int64, error)
}

// Интерфейс репозитория рейтинга
type RatingRepository interface {
	AddSolution(ctx context.Context, userID uuid.UUID, points int, solvedAt time.Time) error
	GetByUserID(ctx context.Context, userID uuid.UUID) (*model.Rating, error)
	List(ctx context.Context, limit, offset int, excludedTelegramIDs []int64) ([]model.Rating, error)
	Count(ctx context.Context, excludedTelegramIDs []int64) (int64, error)
}

// Структура данных ожидающего уведомления
type PendingTaskNotification struct {
	UserID         uuid.UUID // Идентификатор пользователя
	TelegramUserID int64     // Идентификатор пользователя в Telegram
	TaskID         uuid.UUID // Идентификатор таска
}

// Интерфейс репозитория уведомлений
type NotificationRepository interface {
	ListPending(ctx context.Context, now time.Time, limit int) ([]PendingTaskNotification, error)
	MarkSent(ctx context.Context, userID, taskID uuid.UUID, sentAt time.Time) error
}

// Структура набора репозиториев
type Repositories struct {
	Users         UserRepository
	Tasks         TaskRepository
	Submissions   SubmissionRepository
	Ratings       RatingRepository
	Notifications NotificationRepository
}

// Интерфейс выполнения операций в транзакции
type Transactor interface {
	WithinTransaction(ctx context.Context, fn func(Repositories) error) error
}
