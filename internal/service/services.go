package service

import (
	"time"

	"RedButton-bot/internal/repository"
)

// Структура сервисов приложения
type Services struct {
	Users         *UserService         // Сервис пользователей
	Tasks         *TaskService         // Сервис тасков
	Submissions   *SubmissionService   // Сервис попыток сдачи
	Ratings       *RatingService       // Сервис рейтинга
	Notifications *NotificationService // Сервис уведомлений
	Profiles      *ProfileService      // Сервис профиля пользователя
}

// Функция создания сервисов приложения.
// Возвращает готовый набор сервисов для использования в боте.
func New(
	repositories repository.Repositories,
	transactor repository.Transactor,
	verifier FlagVerifier,
	excludedTelegramIDs map[int64]struct{},
	flagSubmissionInterval time.Duration,
) *Services {
	return &Services{
		Users:         NewUserService(repositories.Users),
		Tasks:         NewTaskService(repositories.Tasks),
		Submissions:   NewSubmissionService(transactor, verifier, excludedTelegramIDs, flagSubmissionInterval),
		Ratings:       NewRatingService(repositories.Ratings, excludedTelegramIDs),
		Notifications: NewNotificationService(repositories.Notifications, repositories.Tasks),
		Profiles:      NewProfileService(repositories.Users, repositories.Tasks, repositories.Ratings, excludedTelegramIDs),
	}
}
