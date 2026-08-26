package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"RedButton-bot/internal/database/model"
	"RedButton-bot/internal/dto"
	"github.com/google/uuid"
)

func TestCalculateTaskEndAt(t *testing.T) {
	startsAt := time.Date(2026, time.September, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		expire    time.Duration
		botEndsAt time.Time
		want      time.Time
	}{
		{
			name:      "task expire",
			expire:    24 * time.Hour,
			botEndsAt: startsAt.Add(7 * 24 * time.Hour),
			want:      startsAt.Add(24 * time.Hour),
		},
		{
			name:      "bot end",
			expire:    24 * time.Hour,
			botEndsAt: startsAt.Add(6 * time.Hour),
			want:      startsAt.Add(6 * time.Hour),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := calculateTaskEndAt(startsAt, tt.expire, tt.botEndsAt); !got.Equal(tt.want) {
				t.Errorf("calculateTaskEndAt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTaskSyncMapsDefinitionAndCapsEndTime(t *testing.T) {
	startsAt := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	botEndsAt := startsAt.Add(time.Hour)
	var saved []model.Task
	service := NewTaskService(taskRepositoryStub{sync: func(_ context.Context, tasks []model.Task) error {
		saved = tasks
		return nil
	}})
	err := service.Sync(context.Background(), []dto.TaskDefinition{{
		Slug: "task", Title: "Task", Description: "Description", Flag: "flag",
		MaximumPoints: 100, MinimumPoints: 10, Decay: 5, StartsAt: startsAt,
	}}, 24*time.Hour, startsAt.Add(-time.Hour), botEndsAt)
	if err != nil {
		t.Fatal(err)
	}
	if len(saved) != 1 || saved[0].FlagHash == "flag" || saved[0].EndsAt == nil || !saved[0].EndsAt.Equal(botEndsAt) || !saved[0].IsActive {
		t.Fatalf("unexpected mapped tasks: %#v", saved)
	}
}

func TestTaskServiceValidationAndAvailability(t *testing.T) {
	now := time.Now()
	service := NewTaskService(taskRepositoryStub{})
	if err := service.Sync(context.Background(), nil, time.Hour, now, now.Add(time.Hour)); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("error = %v", err)
	}
	if _, err := service.GetUpcoming(context.Background(), 101); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("error = %v", err)
	}
	if _, err := service.GetNewestForUser(context.Background(), 0); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("error = %v", err)
	}
	if _, err := service.GetByID(context.Background(), uuid.Nil); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("error = %v", err)
	}

	id := uuid.New()
	service = NewTaskService(taskRepositoryStub{get: func(context.Context, uuid.UUID) (*model.Task, error) {
		return &model.Task{ID: id, IsActive: true, StartsAt: now.Add(time.Minute)}, nil
	}})
	service.now = func() time.Time { return now }
	if _, err := service.GetByID(context.Background(), id); !errors.Is(err, ErrTaskUnavailable) {
		t.Fatalf("error = %v", err)
	}
}
