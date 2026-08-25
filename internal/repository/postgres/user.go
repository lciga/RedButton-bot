package postgres

import (
	"context"
	"errors"
	"fmt"

	"RedButton-bot/internal/database/model"
	"RedButton-bot/internal/repository"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Структура репозитория пользователей PostgreSQL
type userRepository struct {
	db *gorm.DB
}

// Функция создания или обновления пользователя.
// Заполняет модель пользователя данными из базы.
func (r *userRepository) Upsert(ctx context.Context, user *model.User) error {
	err := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "telegram_user_id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"telegram_username",
				"first_name",
				"last_name",
				"last_seen_at",
				"updated_at",
			}),
		}).
		Create(user).
		Error
	if err != nil {
		return fmt.Errorf("сохранить пользователя: %w", err)
	}

	return nil
}

// Функция получения пользователя по идентификатору Telegram.
// Возвращает модель пользователя и ошибку.
func (r *userRepository) GetByTelegramID(ctx context.Context, telegramUserID int64) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).
		First(&user, "telegram_user_id = ?", telegramUserID).
		Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("получить пользователя: %w", err)
	}

	return &user, nil
}
