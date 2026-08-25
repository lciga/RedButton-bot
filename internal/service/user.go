package service

import (
	"context"
	"time"

	"RedButton-bot/internal/database/model"
	"RedButton-bot/internal/dto"
	"RedButton-bot/internal/repository"
)

// Структура сервиса пользователей
type UserService struct {
	repository repository.UserRepository
	now        func() time.Time
}

// Функция создания сервиса пользователей.
// Возвращает указатель на экземпляр UserService.
func NewUserService(repository repository.UserRepository) *UserService {
	return &UserService{repository: repository, now: time.Now}
}

// Функция авторизации пользователя через Telegram.
// Создаёт пользователя или обновляет его данные.
func (s *UserService) Authenticate(ctx context.Context, input dto.AuthenticateUser) (*dto.User, error) {
	if input.TelegramUserID <= 0 {
		return nil, ErrInvalidInput
	}

	now := s.now()
	user := model.User{
		TelegramUserID:   input.TelegramUserID,
		TelegramUsername: input.TelegramUsername,
		FirstName:        input.FirstName,
		LastName:         input.LastName,
		IsActive:         true,
		LastSeenAt:       &now,
		UpdatedAt:        now,
	}
	if err := s.repository.Upsert(ctx, &user); err != nil {
		return nil, err
	}

	stored, err := s.repository.GetByTelegramID(ctx, input.TelegramUserID)
	if err != nil {
		return nil, err
	}

	result := mapUser(stored)
	return &result, nil
}
