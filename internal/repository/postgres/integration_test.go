//go:build integration

package postgres_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"RedButton-bot/internal/database"
	"RedButton-bot/internal/database/model"
	"RedButton-bot/internal/repository"
	postgresrepository "RedButton-bot/internal/repository/postgres"
	"RedButton-bot/internal/service"
	"RedButton-bot/internal/testsupport"
)

func TestPostgreSQLRepositoriesAndMigrations(t *testing.T) {
	db := testsupport.PostgreSQL(t)
	if err := database.Migrate(context.Background(), db); err != nil {
		t.Fatalf("idempotent migration: %v", err)
	}
	store := postgresrepository.New(db)
	repositories := store.Repositories()
	now := time.Now().UTC().Truncate(time.Second)

	user := model.User{TelegramUserID: 101, IsActive: true, UpdatedAt: now}
	if err := repositories.Users.Upsert(context.Background(), &user); err != nil {
		t.Fatal(err)
	}
	storedUser, err := repositories.Users.GetByTelegramID(context.Background(), 101)
	if err != nil {
		t.Fatal(err)
	}

	endsAt := now.Add(time.Hour)
	task := model.Task{Slug: "integration", Title: "Integration", Description: "test", FlagHash: service.HashFlag("flag"), InitialPoints: 100, MinimumPoints: 10, CurrentPoints: 100, Decay: 5, StartsAt: now.Add(-time.Minute), EndsAt: &endsAt, IsActive: true}
	if err := repositories.Tasks.Sync(context.Background(), []model.Task{task}); err != nil {
		t.Fatal(err)
	}
	available, err := repositories.Tasks.GetNewestAvailableForUser(context.Background(), 101, now)
	if err != nil {
		t.Fatal(err)
	}
	if available.Slug != "integration" {
		t.Fatalf("task = %#v", available)
	}

	pending, err := repositories.Notifications.ListPending(context.Background(), now, 10)
	if err != nil || len(pending) != 1 {
		t.Fatalf("pending=%#v error=%v", pending, err)
	}
	if err := repositories.Notifications.MarkSent(context.Background(), storedUser.ID, available.ID, now); err != nil {
		t.Fatal(err)
	}
	pending, err = repositories.Notifications.ListPending(context.Background(), now, 10)
	if err != nil || len(pending) != 0 {
		t.Fatalf("pending after mark=%#v error=%v", pending, err)
	}

	rollback := errors.New("rollback")
	err = store.WithinTransaction(context.Background(), func(repositories repository.Repositories) error {
		return repositories.Users.Upsert(context.Background(), &model.User{TelegramUserID: 202, IsActive: true, UpdatedAt: now})
	})
	if err != nil {
		t.Fatal(err)
	}
	err = store.WithinTransaction(context.Background(), func(repositories repository.Repositories) error {
		if err := repositories.Users.Upsert(context.Background(), &model.User{TelegramUserID: 303, IsActive: true, UpdatedAt: now}); err != nil {
			return err
		}
		return rollback
	})
	if !errors.Is(err, rollback) {
		t.Fatalf("transaction error = %v", err)
	}
	if _, err := repositories.Users.GetByTelegramID(context.Background(), 303); !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("rolled back user error = %v", err)
	}
}
