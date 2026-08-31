package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"RedButton-bot/internal/database/model"
	"RedButton-bot/internal/repository"
	"github.com/google/uuid"
)

func TestProfileGet(t *testing.T) {
	userID := uuid.New()
	now := time.Now()
	solvedID, activeID := uuid.New(), uuid.New()
	service := NewProfileService(
		userRepositoryStub{get: func(context.Context, int64) (*model.User, error) {
			return &model.User{ID: userID, TelegramUserID: 42}, nil
		}},
		taskRepositoryStub{profile: func(_ context.Context, id uuid.UUID, at time.Time) ([]repository.ProfileTask, error) {
			if id != userID || !at.Equal(now) {
				t.Fatalf("profile query arguments = %v/%v", id, at)
			}
			return []repository.ProfileTask{{ID: solvedID, Title: "Solved", Solved: true}, {ID: activeID, Title: "Active"}}, nil
		}},
		ratingRepositoryStub{
			get: func(context.Context, uuid.UUID) (*model.Rating, error) {
				return &model.Rating{TotalPoints: 150}, nil
			},
			position: func(context.Context, uuid.UUID, []int64) (int, error) { return 3, nil },
		},
		nil,
	)
	service.now = func() time.Time { return now }
	profile, err := service.Get(context.Background(), 42)
	if err != nil {
		t.Fatal(err)
	}
	if profile.TotalPoints != 150 || profile.Position != 3 || len(profile.Tasks) != 2 || !profile.Tasks[0].Solved {
		t.Fatalf("profile = %#v", profile)
	}
}

func TestProfileWithoutRatingAndValidation(t *testing.T) {
	service := NewProfileService(userRepositoryStub{}, taskRepositoryStub{}, ratingRepositoryStub{}, nil)
	if _, err := service.Get(context.Background(), 0); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("error = %v, want ErrInvalidInput", err)
	}

	userID := uuid.New()
	service = NewProfileService(
		userRepositoryStub{get: func(context.Context, int64) (*model.User, error) { return &model.User{ID: userID}, nil }},
		taskRepositoryStub{profile: func(context.Context, uuid.UUID, time.Time) ([]repository.ProfileTask, error) { return nil, nil }},
		ratingRepositoryStub{get: func(context.Context, uuid.UUID) (*model.Rating, error) { return nil, repository.ErrNotFound }},
		nil,
	)
	profile, err := service.Get(context.Background(), 42)
	if err != nil || profile.TotalPoints != 0 || profile.Position != 0 {
		t.Fatalf("profile=%#v error=%v", profile, err)
	}
}
