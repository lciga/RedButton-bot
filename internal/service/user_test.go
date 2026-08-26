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

func TestUserAuthenticate(t *testing.T) {
	now := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	id := uuid.New()
	var saved *model.User
	repository := userRepositoryStub{
		upsert: func(_ context.Context, user *model.User) error { saved = user; return nil },
		get: func(_ context.Context, telegramID int64) (*model.User, error) {
			return &model.User{ID: id, TelegramUserID: telegramID, IsActive: true, LastSeenAt: &now}, nil
		},
	}
	service := NewUserService(repository)
	service.now = func() time.Time { return now }

	got, err := service.Authenticate(context.Background(), dto.AuthenticateUser{TelegramUserID: 42})
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != id || !got.IsActive || saved == nil || saved.TelegramUserID != 42 || !saved.UpdatedAt.Equal(now) {
		t.Fatalf("unexpected authentication result: got=%#v saved=%#v", got, saved)
	}
}

func TestUserAuthenticateValidationAndRepositoryError(t *testing.T) {
	service := NewUserService(userRepositoryStub{})
	if _, err := service.Authenticate(context.Background(), dto.AuthenticateUser{}); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("error = %v, want ErrInvalidInput", err)
	}
	wantErr := errors.New("database unavailable")
	service = NewUserService(userRepositoryStub{upsert: func(context.Context, *model.User) error { return wantErr }})
	if _, err := service.Authenticate(context.Background(), dto.AuthenticateUser{TelegramUserID: 1}); !errors.Is(err, wantErr) {
		t.Fatalf("error = %v, want %v", err, wantErr)
	}
}
