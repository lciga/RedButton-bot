package postgres

import (
	"context"
	"fmt"
	"time"

	"RedButton-bot/internal/database/model"
	"RedButton-bot/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Структура репозитория уведомлений PostgreSQL
type notificationRepository struct {
	db *gorm.DB
}

// Функция получения неотправленных уведомлений.
// Возвращает активных пользователей и доступные им новые таски.
func (r *notificationRepository) ListPending(
	ctx context.Context,
	now time.Time,
	limit int,
) ([]repository.PendingTaskNotification, error) {
	var notifications []repository.PendingTaskNotification
	err := r.db.WithContext(ctx).
		Table("users AS users").
		Select("users.id AS user_id, users.telegram_user_id, tasks.id AS task_id").
		Joins("CROSS JOIN tasks AS tasks").
		Where("users.is_active = ?", true).
		Where("tasks.is_active = ?", true).
		Where("tasks.starts_at <= ?", now).
		Where("tasks.ends_at IS NULL OR tasks.ends_at > ?", now).
		Where(`NOT EXISTS (
			SELECT 1
			FROM task_notifications
			WHERE task_notifications.user_id = users.id
			  AND task_notifications.task_id = tasks.id
		)`).
		Order("tasks.starts_at ASC, users.id ASC").
		Limit(limit).
		Scan(&notifications).
		Error
	if err != nil {
		return nil, fmt.Errorf("list notifications: %w", err)
	}

	return notifications, nil
}

// Функция сохранения отправленного уведомления.
// Повторный вызов для пользователя и таска игнорируется.
func (r *notificationRepository) MarkSent(
	ctx context.Context,
	userID, taskID uuid.UUID,
	sentAt time.Time,
) error {
	notification := model.TaskNotification{
		UserID: userID,
		TaskID: taskID,
		SentAt: sentAt,
	}
	err := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&notification).
		Error
	if err != nil {
		return fmt.Errorf("save notification: %w", err)
	}

	return nil
}
