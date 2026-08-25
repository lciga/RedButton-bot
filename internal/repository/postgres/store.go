package postgres

import (
	"context"
	"fmt"

	"RedButton-bot/internal/repository"
	"gorm.io/gorm"
)

// Структура хранилища PostgreSQL
type Store struct {
	db *gorm.DB
}

// Функция создания хранилища PostgreSQL.
// Возвращает указатель на экземпляр Store.
func New(db *gorm.DB) *Store {
	return &Store{db: db}
}

// Функция получения набора репозиториев.
// Возвращает репозитории с общим соединением GORM.
func (s *Store) Repositories() repository.Repositories {
	return newRepositories(s.db)
}

// Функция выполнения операций внутри транзакции.
// Передаёт в callback репозитории с транзакционным соединением.
func (s *Store) WithinTransaction(ctx context.Context, fn func(repository.Repositories) error) error {
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(newRepositories(tx))
	})
	if err != nil {
		return fmt.Errorf("выполнить транзакцию: %w", err)
	}

	return nil
}

func newRepositories(db *gorm.DB) repository.Repositories {
	return repository.Repositories{
		Users:         &userRepository{db: db},
		Tasks:         &taskRepository{db: db},
		Submissions:   &submissionRepository{db: db},
		Ratings:       &ratingRepository{db: db},
		Notifications: &notificationRepository{db: db},
	}
}
