package service

import (
	"context"
	"fmt"
	"time"

	"RedButton-bot/internal/database/model"
	"RedButton-bot/internal/dto"
	"RedButton-bot/internal/repository"
	"github.com/google/uuid"
)

// Структура сервиса тасков
type TaskService struct {
	repository repository.TaskRepository
	now        func() time.Time
}

// Функция создания сервиса тасков.
// Возвращает указатель на экземпляр TaskService.
func NewTaskService(repository repository.TaskRepository) *TaskService {
	return &TaskService{repository: repository, now: time.Now}
}

// Функция синхронизации тасков из YAML-файлов.
// Вычисляет время закрытия на основании TASK_EXPIRE.
func (s *TaskService) Sync(
	ctx context.Context,
	definitions []dto.TaskDefinition,
	expire time.Duration,
	botStartsAt, botEndsAt time.Time,
) error {
	if len(definitions) == 0 || expire <= 0 || !botEndsAt.After(botStartsAt) {
		return ErrInvalidInput
	}

	tasks := make([]model.Task, 0, len(definitions))
	for _, definition := range definitions {
		if definition.StartsAt.Before(botStartsAt) || !definition.StartsAt.Before(botEndsAt) {
			return fmt.Errorf(
				"%w: время открытия таска %q находится вне периода работы бота",
				ErrInvalidInput,
				definition.Slug,
			)
		}
		endsAt := calculateTaskEndAt(definition.StartsAt, expire, botEndsAt)
		task := model.Task{
			Slug:          definition.Slug,
			Title:         definition.Title,
			Description:   definition.Description,
			FlagHash:      HashFlag(definition.Flag),
			InitialPoints: definition.MaximumPoints,
			MinimumPoints: definition.MinimumPoints,
			CurrentPoints: definition.MaximumPoints,
			Decay:         definition.Decay,
			StartsAt:      definition.StartsAt,
			EndsAt:        &endsAt,
			IsActive:      true,
		}
		if definition.File != nil {
			task.Files = []model.TaskFile{
				{
					StoragePath: definition.File.StoragePath,
					FileName:    definition.File.FileName,
					MIMEType:    definition.File.MIMEType,
					FileSize:    definition.File.FileSize,
				},
			}
		}
		tasks = append(tasks, task)
	}

	return s.repository.Sync(ctx, tasks)
}

func calculateTaskEndAt(startsAt time.Time, expire time.Duration, botEndsAt time.Time) time.Time {
	endsAt := startsAt.Add(expire)
	if endsAt.After(botEndsAt) {
		return botEndsAt
	}

	return endsAt
}

// Функция получения доступных тасков.
// Возвращает безопасные данные без хэша флага.
func (s *TaskService) GetAvailable(ctx context.Context) ([]dto.Task, error) {
	tasks, err := s.repository.ListAvailable(ctx, s.now())
	if err != nil {
		return nil, err
	}

	result := make([]dto.Task, 0, len(tasks))
	for i := range tasks {
		result = append(result, mapTask(&tasks[i]))
	}

	return result, nil
}

// Функция получения предстоящих тасков.
// Возвращает безопасные данные без хэшей флагов.
func (s *TaskService) GetUpcoming(ctx context.Context, limit int) ([]dto.Task, error) {
	if limit <= 0 || limit > 100 {
		return nil, ErrInvalidInput
	}

	tasks, err := s.repository.ListUpcoming(ctx, s.now(), limit)
	if err != nil {
		return nil, err
	}

	result := make([]dto.Task, 0, len(tasks))
	for i := range tasks {
		result = append(result, mapTask(&tasks[i]))
	}

	return result, nil
}

// Функция получения новой таски для пользователя.
// Возвращает последнюю доступную и ещё не решённую таску.
func (s *TaskService) GetNewestForUser(ctx context.Context, telegramUserID int64) (*dto.Task, error) {
	if telegramUserID <= 0 {
		return nil, ErrInvalidInput
	}

	task, err := s.repository.GetNewestAvailableForUser(ctx, telegramUserID, s.now())
	if err != nil {
		return nil, err
	}

	result := mapTask(task)
	return &result, nil
}

// Функция получения таска по идентификатору.
// Возвращает безопасные данные без хэша флага.
func (s *TaskService) GetByID(ctx context.Context, id uuid.UUID) (*dto.Task, error) {
	if id == uuid.Nil {
		return nil, ErrInvalidInput
	}

	task, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if !task.IsActive || task.StartsAt.After(s.now()) || task.EndsAt != nil && !task.EndsAt.After(s.now()) {
		return nil, ErrTaskUnavailable
	}

	result := mapTask(task)
	return &result, nil
}
