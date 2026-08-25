package service

import (
	"context"

	"RedButton-bot/internal/dto"
	"RedButton-bot/internal/repository"
	"github.com/google/uuid"
)

// Структура сервиса рейтинга
type RatingService struct {
	repository          repository.RatingRepository
	excludedTelegramIDs []int64
}

// Функция создания сервиса рейтинга.
// Возвращает указатель на экземпляр RatingService.
func NewRatingService(
	repository repository.RatingRepository,
	excludedTelegramIDs map[int64]struct{},
) *RatingService {
	excluded := make([]int64, 0, len(excludedTelegramIDs))
	for telegramID := range excludedTelegramIDs {
		excluded = append(excluded, telegramID)
	}

	return &RatingService{
		repository:          repository,
		excludedTelegramIDs: excluded,
	}
}

// Функция получения таблицы рейтинга.
// Возвращает страницу рейтинга с позициями пользователей.
func (s *RatingService) GetLeaderboard(ctx context.Context, limit, offset int) ([]dto.Rating, error) {
	if limit <= 0 || limit > 100 || offset < 0 {
		return nil, ErrInvalidInput
	}

	ratings, err := s.repository.List(ctx, limit, offset, s.excludedTelegramIDs)
	if err != nil {
		return nil, err
	}

	result := make([]dto.Rating, 0, len(ratings))
	for i := range ratings {
		result = append(result, mapRating(&ratings[i], offset+i+1))
	}

	return result, nil
}

// Функция получения страницы рейтинга.
// Возвращает участников и параметры пагинации.
func (s *RatingService) GetLeaderboardPage(
	ctx context.Context,
	page, pageSize int,
) (*dto.RatingPage, error) {
	if page <= 0 || pageSize <= 0 || pageSize > 100 {
		return nil, ErrInvalidInput
	}

	totalItems, err := s.repository.Count(ctx, s.excludedTelegramIDs)
	if err != nil {
		return nil, err
	}
	totalPages := int((totalItems + int64(pageSize) - 1) / int64(pageSize))
	if totalPages == 0 {
		totalPages = 1
	}
	if page > totalPages {
		return nil, ErrInvalidInput
	}

	offset := (page - 1) * pageSize
	items, err := s.GetLeaderboard(ctx, pageSize, offset)
	if err != nil {
		return nil, err
	}

	return &dto.RatingPage{
		Items:      items,
		Page:       page,
		TotalPages: totalPages,
		TotalItems: totalItems,
	}, nil
}

// Функция получения рейтинга пользователя.
// Возвращает данные рейтинга без вычисления общей позиции.
func (s *RatingService) GetByUserID(ctx context.Context, userID uuid.UUID) (*dto.Rating, error) {
	if userID == uuid.Nil {
		return nil, ErrInvalidInput
	}

	rating, err := s.repository.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	for _, telegramID := range s.excludedTelegramIDs {
		if rating.User.TelegramUserID == telegramID {
			return nil, repository.ErrNotFound
		}
	}

	result := mapRating(rating, 0)
	return &result, nil
}
