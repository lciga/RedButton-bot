package service

import "RedButton-bot/internal/repository"

// Структура сервисов приложения
type Services struct {
	Users         *UserService         // Сервис пользователей
	Tasks         *TaskService         // Сервис тасков
	Submissions   *SubmissionService   // Сервис попыток сдачи
	Ratings       *RatingService       // Сервис рейтинга
	Notifications *NotificationService // Сервис уведомлений
}

// Функция создания сервисов приложения.
// Возвращает готовый набор сервисов для использования в боте.
func New(
	repositories repository.Repositories,
	transactor repository.Transactor,
	verifier FlagVerifier,
	excludedTelegramIDs map[int64]struct{},
) *Services {
	return &Services{
		Users:         NewUserService(repositories.Users),
		Tasks:         NewTaskService(repositories.Tasks),
		Submissions:   NewSubmissionService(transactor, verifier, excludedTelegramIDs),
		Ratings:       NewRatingService(repositories.Ratings, excludedTelegramIDs),
		Notifications: NewNotificationService(repositories.Notifications, repositories.Tasks),
	}
}
