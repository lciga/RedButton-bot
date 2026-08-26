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

func TestNotificationGetPendingCachesTasks(t *testing.T) {
	taskID, userOne, userTwo := uuid.New(), uuid.New(), uuid.New()
	getCalls := 0
	notifications := notificationRepositoryStub{list: func(context.Context, time.Time, int) ([]repository.PendingTaskNotification, error) {
		return []repository.PendingTaskNotification{{UserID: userOne, TelegramUserID: 1, TaskID: taskID}, {UserID: userTwo, TelegramUserID: 2, TaskID: taskID}}, nil
	}}
	tasks := taskRepositoryStub{get: func(context.Context, uuid.UUID) (*model.Task, error) {
		getCalls++
		return &model.Task{ID: taskID, Title: "Task"}, nil
	}}
	service := NewNotificationService(notifications, tasks)
	got, err := service.GetPending(context.Background(), 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || getCalls != 1 {
		t.Fatalf("notifications=%d task lookups=%d, want 2/1", len(got), getCalls)
	}
}

func TestNotificationValidationAndMarkSent(t *testing.T) {
	service := NewNotificationService(notificationRepositoryStub{}, taskRepositoryStub{})
	if _, err := service.GetPending(context.Background(), 0); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("error = %v", err)
	}
	if err := service.MarkSent(context.Background(), uuid.Nil, uuid.New()); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("error = %v", err)
	}

	now := time.Now()
	called := false
	service = NewNotificationService(notificationRepositoryStub{mark: func(_ context.Context, _, _ uuid.UUID, at time.Time) error {
		called = at.Equal(now)
		return nil
	}}, taskRepositoryStub{})
	service.now = func() time.Time { return now }
	if err := service.MarkSent(context.Background(), uuid.New(), uuid.New()); err != nil || !called {
		t.Fatalf("error=%v called=%v", err, called)
	}
}
