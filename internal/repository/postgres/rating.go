package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"RedButton-bot/internal/database/model"
	"RedButton-bot/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Структура репозитория рейтинга PostgreSQL
type ratingRepository struct {
	db *gorm.DB
}

// Функция добавления решённого таска в рейтинг.
// Создаёт рейтинг пользователя при первом решении.
func (r *ratingRepository) AddSolution(ctx context.Context, userID uuid.UUID, points int, solvedAt time.Time) error {
	rating := model.Rating{
		UserID:           userID,
		TotalPoints:      points,
		SolvedTasksCount: 1,
		LastSolvedAt:     &solvedAt,
		UpdatedAt:        solvedAt,
	}

	err := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "user_id"}},
			DoUpdates: clause.Assignments(map[string]any{
				"total_points":       gorm.Expr("ratings.total_points + ?", points),
				"solved_tasks_count": gorm.Expr("ratings.solved_tasks_count + 1"),
				"last_solved_at":     solvedAt,
				"updated_at":         solvedAt,
			}),
		}).
		Create(&rating).
		Error
	if err != nil {
		return fmt.Errorf("обновить рейтинг: %w", err)
	}

	return nil
}

// Функция получения рейтинга пользователя.
// Возвращает рейтинг вместе с пользователем.
func (r *ratingRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*model.Rating, error) {
	var rating model.Rating
	err := r.db.WithContext(ctx).
		Preload("User").
		Where("solved_tasks_count > 0").
		First(&rating, "user_id = ?", userID).
		Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("получить рейтинг пользователя: %w", err)
	}

	return &rating, nil
}

// Функция получения таблицы рейтинга.
// Возвращает пользователей в порядке количества очков.
func (r *ratingRepository) List(
	ctx context.Context,
	limit, offset int,
	excludedTelegramIDs []int64,
) ([]model.Rating, error) {
	var ratings []model.Rating
	query := r.ratingQuery(ctx, excludedTelegramIDs).
		Select("ratings.*").
		Preload("User").
		Order("ratings.total_points DESC, ratings.last_solved_at ASC NULLS LAST, ratings.user_id ASC").
		Limit(limit).
		Offset(offset)
	err := query.Find(&ratings).Error
	if err != nil {
		return nil, fmt.Errorf("получить таблицу рейтинга: %w", err)
	}

	return ratings, nil
}

// Функция подсчёта участников рейтинга.
// Учитывает только пользователей с правильными решениями.
func (r *ratingRepository) Count(ctx context.Context, excludedTelegramIDs []int64) (int64, error) {
	var count int64
	if err := r.ratingQuery(ctx, excludedTelegramIDs).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("посчитать участников рейтинга: %w", err)
	}

	return count, nil
}

func (r *ratingRepository) ratingQuery(ctx context.Context, excludedTelegramIDs []int64) *gorm.DB {
	query := r.db.WithContext(ctx).
		Model(&model.Rating{}).
		Where("ratings.solved_tasks_count > 0")
	if len(excludedTelegramIDs) != 0 {
		query = query.
			Joins("JOIN users ON users.id = ratings.user_id").
			Where("users.telegram_user_id NOT IN ?", excludedTelegramIDs)
	}

	return query
}
