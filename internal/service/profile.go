package service

import (
	"context"
	"errors"
	"time"

	"RedButton-bot/internal/dto"
	"RedButton-bot/internal/repository"
)

// Структура сервиса профиля пользователя.
type ProfileService struct {
	users               repository.UserRepository
	tasks               repository.TaskRepository
	ratings             repository.RatingRepository
	excludedTelegramIDs []int64
	now                 func() time.Time
}

// Функция создания сервиса профиля пользователя.
func NewProfileService(
	users repository.UserRepository,
	tasks repository.TaskRepository,
	ratings repository.RatingRepository,
	excludedTelegramIDs map[int64]struct{},
) *ProfileService {
	excluded := make([]int64, 0, len(excludedTelegramIDs))
	for telegramID := range excludedTelegramIDs {
		excluded = append(excluded, telegramID)
	}
	return &ProfileService{
		users: users, tasks: tasks, ratings: ratings,
		excludedTelegramIDs: excluded,
		now:                 time.Now,
	}
}

// Функция получения очков, места и списка задач пользователя.
func (s *ProfileService) Get(ctx context.Context, telegramUserID int64) (*dto.Profile, error) {
	if telegramUserID <= 0 {
		return nil, ErrInvalidInput
	}
	user, err := s.users.GetByTelegramID(ctx, telegramUserID)
	if err != nil {
		return nil, err
	}

	result := &dto.Profile{}
	rating, err := s.ratings.GetByUserID(ctx, user.ID)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}
	if err == nil {
		result.TotalPoints = rating.TotalPoints
		position, err := s.ratings.GetPosition(ctx, user.ID, s.excludedTelegramIDs)
		if err != nil && !errors.Is(err, repository.ErrNotFound) {
			return nil, err
		}
		result.Position = position
	}

	tasks, err := s.tasks.ListForProfile(ctx, user.ID, s.now())
	if err != nil {
		return nil, err
	}
	result.Tasks = make([]dto.ProfileTask, 0, len(tasks))
	for _, task := range tasks {
		result.Tasks = append(result.Tasks, dto.ProfileTask{ID: task.ID, Title: task.Title, Solved: task.Solved})
	}
	return result, nil
}
