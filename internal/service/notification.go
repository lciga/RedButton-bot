package service

import (
	"context"
	"time"

	"RedButton-bot/internal/dto"
	"RedButton-bot/internal/repository"
	"github.com/google/uuid"
)

// Структура сервиса уведомлений
type NotificationService struct {
	notifications repository.NotificationRepository
	tasks         repository.TaskRepository
	now           func() time.Time
}

// Функция создания сервиса уведомлений.
// Возвращает указатель на экземпляр NotificationService.
func NewNotificationService(
	notifications repository.NotificationRepository,
	tasks repository.TaskRepository,
) *NotificationService {
	return &NotificationService{
		notifications: notifications,
		tasks:         tasks,
		now:           time.Now,
	}
}

// Функция получения ожидающих уведомлений.
// Возвращает готовые данные пользователей и тасков.
func (s *NotificationService) GetPending(ctx context.Context, limit int) ([]dto.TaskNotification, error) {
	if limit <= 0 || limit > 1000 {
		return nil, ErrInvalidInput
	}

	pending, err := s.notifications.ListPending(ctx, s.now(), limit)
	if err != nil {
		return nil, err
	}

	tasks := make(map[uuid.UUID]dto.Task)
	result := make([]dto.TaskNotification, 0, len(pending))
	for _, notification := range pending {
		task, exists := tasks[notification.TaskID]
		if !exists {
			stored, err := s.tasks.GetByID(ctx, notification.TaskID)
			if err != nil {
				return nil, err
			}
			task = mapTask(stored)
			tasks[notification.TaskID] = task
		}

		result = append(result, dto.TaskNotification{
			UserID:         notification.UserID,
			TelegramUserID: notification.TelegramUserID,
			Task:           task,
		})
	}

	return result, nil
}

// Функция получения времени открытия ближайшего таска.
// Используется для точного запуска рассылки по starts_at.
func (s *NotificationService) GetNextStartsAt(ctx context.Context) (*time.Time, error) {
	return s.tasks.GetNextStartsAt(ctx, s.now())
}

// Функция подтверждения отправки уведомления.
// Сохраняет время успешной отправки.
func (s *NotificationService) MarkSent(ctx context.Context, userID, taskID uuid.UUID) error {
	if userID == uuid.Nil || taskID == uuid.Nil {
		return ErrInvalidInput
	}

	return s.notifications.MarkSent(ctx, userID, taskID, s.now())
}
