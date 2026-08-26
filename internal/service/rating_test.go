package service

import (
	"context"
	"errors"
	"slices"
	"testing"

	"RedButton-bot/internal/database/model"
	"RedButton-bot/internal/repository"
	"github.com/google/uuid"
)

func TestRatingLeaderboardPage(t *testing.T) {
	username := "winner"
	repository := ratingRepositoryStub{
		count: func(_ context.Context, excluded []int64) (int64, error) {
			if !slices.Contains(excluded, int64(999)) {
				t.Error("administrator was not excluded")
			}
			return 11, nil
		},
		list: func(_ context.Context, limit, offset int, _ []int64) ([]model.Rating, error) {
			if limit != 10 || offset != 10 {
				t.Fatalf("pagination = %d/%d, want 10/10", limit, offset)
			}
			return []model.Rating{{UserID: uuid.New(), TotalPoints: 100, SolvedTasksCount: 1, User: model.User{TelegramUserID: 42, TelegramUsername: &username}}}, nil
		},
	}
	service := NewRatingService(repository, map[int64]struct{}{999: {}})
	page, err := service.GetLeaderboardPage(context.Background(), 2, 10)
	if err != nil {
		t.Fatal(err)
	}
	if page.Page != 2 || page.TotalPages != 2 || page.TotalItems != 11 || len(page.Items) != 1 || page.Items[0].Position != 11 {
		t.Fatalf("unexpected page: %#v", page)
	}
}

func TestRatingValidationAndExcludedUser(t *testing.T) {
	service := NewRatingService(ratingRepositoryStub{}, nil)
	if _, err := service.GetLeaderboardPage(context.Background(), 0, 10); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("error = %v", err)
	}
	if _, err := service.GetByUserID(context.Background(), uuid.Nil); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("error = %v", err)
	}

	id := uuid.New()
	service = NewRatingService(ratingRepositoryStub{get: func(context.Context, uuid.UUID) (*model.Rating, error) {
		return &model.Rating{User: model.User{TelegramUserID: 999}}, nil
	}}, map[int64]struct{}{999: {}})
	if _, err := service.GetByUserID(context.Background(), id); !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("error = %v", err)
	}
}
