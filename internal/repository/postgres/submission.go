package postgres

import (
	"context"
	"fmt"

	"RedButton-bot/internal/database/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Структура репозитория попыток сдачи PostgreSQL
type submissionRepository struct {
	db *gorm.DB
}

// Функция сохранения попытки сдачи.
// Возвращает ошибку сохранения.
func (r *submissionRepository) Create(ctx context.Context, submission *model.Submission) error {
	if err := r.db.WithContext(ctx).Create(submission).Error; err != nil {
		return fmt.Errorf("save submission: %w", err)
	}

	return nil
}

// Функция проверки решённого таска пользователем.
// Возвращает true при наличии правильного решения.
func (r *submissionRepository) HasCorrect(ctx context.Context, userID, taskID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.Submission{}).
		Where("user_id = ? AND task_id = ? AND is_correct = ?", userID, taskID, true).
		Count(&count).
		Error
	if err != nil {
		return false, fmt.Errorf("check task solution: %w", err)
	}

	return count != 0, nil
}

// Функция подсчёта правильных решений таска.
// Возвращает количество решений и ошибку.
func (r *submissionRepository) CountCorrect(ctx context.Context, taskID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.Submission{}).
		Where("task_id = ? AND is_correct = ?", taskID, true).
		Count(&count).
		Error
	if err != nil {
		return 0, fmt.Errorf("count task solutions: %w", err)
	}

	return count, nil
}
