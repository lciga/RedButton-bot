package bot

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"RedButton-bot/internal/service"
	telegram "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/google/uuid"
)

const (
	buttonNewTask       = "🧩 Новая таска"
	buttonRating        = "🏆 Рейтинг"
	buttonAdmin         = "⚙️ Админка"
	solvePrefix         = "solve:"
	ratingPrefix        = "rating:"
	adminPreviewPrefix  = "admin:preview:"
	telegramPollTimeout = time.Minute
	telegramHTTPTimeout = 90 * time.Second
)

// Структура Telegram-бота
type Bot struct {
	client               *telegram.Bot      // Клиент Telegram Bot API
	services             *service.Services  // Сервисы приложения
	logger               *slog.Logger       // Логгер приложения
	notificationInterval time.Duration      // Интервал проверки новых тасков
	pendingTasks         sync.Map           // Таски, ожидающие флаг пользователя
	adminTelegramIDs     map[int64]struct{} // Идентификаторы администраторов в Telegram
	startsAt             time.Time          // Время начала работы бота
	endsAt               time.Time          // Время окончания работы бота
}

// Функция создания Telegram-бота.
// Возвращает готовый экземпляр Bot и ошибку инициализации.
func New(
	token string,
	services *service.Services,
	logger *slog.Logger,
	notificationInterval time.Duration,
	telegramInitTimeout time.Duration,
	adminTelegramIDs map[int64]struct{},
	startsAt, endsAt time.Time,
) (*Bot, error) {
	if services == nil {
		return nil, fmt.Errorf("создать Telegram-бота: сервисы не заданы")
	}
	if logger == nil {
		logger = slog.Default()
	}
	if notificationInterval <= 0 {
		notificationInterval = 15 * time.Second
	}
	if telegramInitTimeout <= 0 {
		telegramInitTimeout = 30 * time.Second
	}

	admins := make(map[int64]struct{}, len(adminTelegramIDs))
	for telegramID := range adminTelegramIDs {
		admins[telegramID] = struct{}{}
	}

	application := &Bot{
		services:             services,
		logger:               logger,
		notificationInterval: notificationInterval,
		adminTelegramIDs:     admins,
		startsAt:             startsAt,
		endsAt:               endsAt,
	}
	client, err := telegram.New(
		token,
		telegram.WithCheckInitTimeout(telegramInitTimeout),
		telegram.WithHTTPClient(telegramPollTimeout, newHTTPClient()),
		telegram.WithDefaultHandler(application.handleUpdate),
		telegram.WithErrorsHandler(func(err error) {
			logger.Error("ошибка Telegram Bot API", "error", err)
		}),
		telegram.WithAllowedUpdates(telegram.AllowedUpdates{
			models.AllowedUpdateMessage,
			models.AllowedUpdateCallbackQuery,
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("создать клиент Telegram: %w", err)
	}
	application.client = client

	return application, nil
}

func newHTTPClient() *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.ForceAttemptHTTP2 = false
	transport.TLSNextProto = make(map[string]func(string, *tls.Conn) http.RoundTripper)
	transport.ProxyConnectHeader = http.Header{
		"Proxy-Connection": {"Keep-Alive"},
		"User-Agent":       {"RedButton-bot/1.0"},
	}

	return &http.Client{
		Transport: transport,
		Timeout:   telegramHTTPTimeout,
	}
}

func (b *Bot) isAdmin(telegramUserID int64) bool {
	_, exists := b.adminTelegramIDs[telegramUserID]
	return exists
}

func (b *Bot) isAvailable(now time.Time) bool {
	return !now.Before(b.startsAt) && now.Before(b.endsAt)
}

// Функция запуска Telegram-бота.
// Блокирует выполнение до завершения контекста.
func (b *Bot) Start(ctx context.Context) {
	go b.runNotifications(ctx)
	b.client.Start(ctx)
}

func (b *Bot) setPendingTask(telegramUserID int64, taskID uuid.UUID) {
	b.pendingTasks.Store(telegramUserID, taskID)
}

func (b *Bot) getPendingTask(telegramUserID int64) (uuid.UUID, bool) {
	value, exists := b.pendingTasks.Load(telegramUserID)
	if !exists {
		return uuid.Nil, false
	}

	taskID, ok := value.(uuid.UUID)
	return taskID, ok
}

func (b *Bot) deletePendingTask(telegramUserID int64) {
	b.pendingTasks.Delete(telegramUserID)
}
