package bot

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"RedButton-bot/internal/dto"
	"RedButton-bot/internal/repository"
	"RedButton-bot/internal/service"
	telegram "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/google/uuid"
)

func (b *Bot) handleUpdate(ctx context.Context, client *telegram.Bot, update *models.Update) {
	if update.CallbackQuery != nil {
		b.handleCallback(ctx, client, update)
		return
	}
	if update.Message == nil || update.Message.From == nil || update.Message.From.IsBot {
		return
	}

	user, err := b.authenticate(ctx, update.Message.From)
	if err != nil {
		b.handleError(ctx, update.Message.Chat.ID, err)
		return
	}
	if !user.IsActive {
		b.sendText(ctx, update.Message.Chat.ID, "Ваш аккаунт заблокирован.", nil)
		return
	}
	if !b.isAdmin(update.Message.From.ID) && !b.isAvailable(time.Now()) {
		b.sendUnavailable(ctx, update.Message.Chat.ID)
		return
	}

	switch update.Message.Text {
	case "/start":
		b.handleStart(ctx, update.Message.Chat.ID, update.Message.From.ID)
	case buttonNewTask:
		b.handleNewTask(ctx, update.Message.Chat.ID, update.Message.From.ID)
	case buttonRating:
		b.handleRating(ctx, update.Message.Chat.ID, 1)
	case buttonAdmin:
		b.handleAdmin(ctx, update.Message.Chat.ID, update.Message.From.ID)
	default:
		b.handleFlag(ctx, update.Message.Chat.ID, update.Message.From.ID, update.Message.Text)
	}
}

func (b *Bot) handleStart(ctx context.Context, chatID, telegramUserID int64) {
	b.sendText(
		ctx,
		chatID,
		"Добро пожаловать! Выберите действие в меню.",
		b.mainMenu(telegramUserID),
	)
}

func (b *Bot) handleNewTask(ctx context.Context, chatID, telegramUserID int64) {
	task, err := b.services.Tasks.GetNewestForUser(ctx, telegramUserID)
	if errors.Is(err, repository.ErrNotFound) {
		b.sendText(ctx, chatID, "У вас нет новых доступных тасков.", b.mainMenu(telegramUserID))
		return
	}
	if err != nil {
		b.handleError(ctx, chatID, err)
		return
	}

	if err := b.sendTask(ctx, chatID, *task, false); err != nil {
		b.handleError(ctx, chatID, err)
	}
}

func (b *Bot) handleRating(ctx context.Context, chatID int64, page int) {
	rating, err := b.services.Ratings.GetLeaderboardPage(ctx, page, 10)
	if err != nil {
		b.handleError(ctx, chatID, err)
		return
	}

	b.sendText(ctx, chatID, formatRating(*rating), ratingKeyboard(*rating))
}

func (b *Bot) handleCallback(ctx context.Context, client *telegram.Bot, update *models.Update) {
	query := update.CallbackQuery
	user, err := b.authenticate(ctx, &query.From)
	if err != nil {
		b.answerCallback(ctx, client, query.ID, "Не удалось проверить пользователя.", true)
		b.handleError(ctx, query.From.ID, err)
		return
	}
	if !user.IsActive {
		b.answerCallback(ctx, client, query.ID, "Ваш аккаунт заблокирован.", true)
		b.sendText(ctx, query.From.ID, "Ваш аккаунт заблокирован.", nil)
		return
	}
	if !b.isAdmin(query.From.ID) && !b.isAvailable(time.Now()) {
		b.answerCallback(ctx, client, query.ID, "Бот сейчас недоступен.", true)
		b.sendUnavailable(ctx, query.From.ID)
		return
	}

	switch {
	case strings.HasPrefix(query.Data, solvePrefix):
		b.handleSolveCallback(ctx, client, query)
	case strings.HasPrefix(query.Data, ratingPrefix):
		b.handleRatingCallback(ctx, client, query)
	case strings.HasPrefix(query.Data, adminPreviewPrefix):
		b.handleAdminPreviewCallback(ctx, client, query)
	default:
		b.answerCallback(ctx, client, query.ID, "", false)
	}
}

func (b *Bot) handleRatingCallback(ctx context.Context, client *telegram.Bot, query *models.CallbackQuery) {
	page, err := strconv.Atoi(strings.TrimPrefix(query.Data, ratingPrefix))
	if err != nil || page <= 0 {
		b.answerCallback(ctx, client, query.ID, "Некорректная страница.", true)
		return
	}

	b.answerCallback(ctx, client, query.ID, "", false)
	rating, err := b.services.Ratings.GetLeaderboardPage(ctx, page, 10)
	if err != nil {
		b.handleError(ctx, query.From.ID, err)
		return
	}
	if query.Message.Message == nil {
		b.sendText(ctx, query.From.ID, formatRating(*rating), ratingKeyboard(*rating))
		return
	}

	if _, err := client.EditMessageText(ctx, &telegram.EditMessageTextParams{
		ChatID:      query.Message.Message.Chat.ID,
		MessageID:   query.Message.Message.ID,
		Text:        formatRating(*rating),
		ReplyMarkup: ratingKeyboard(*rating),
	}); err != nil {
		b.handleError(ctx, query.From.ID, fmt.Errorf("update leaderboard page: %w", err))
	}
}

func (b *Bot) handleSolveCallback(ctx context.Context, client *telegram.Bot, query *models.CallbackQuery) {
	taskID, err := uuid.Parse(strings.TrimPrefix(query.Data, solvePrefix))
	if err != nil {
		b.answerCallback(ctx, client, query.ID, "Некорректный таск.", true)
		return
	}

	b.answerCallback(ctx, client, query.ID, "", false)
	b.setPendingTask(query.From.ID, taskID)
	b.sendText(
		ctx,
		query.From.ID,
		"Отправьте флаг следующим сообщением.",
		&models.ForceReply{
			ForceReply:            true,
			InputFieldPlaceholder: "redbutton{flag}",
		},
	)
}

func (b *Bot) handleAdminPreviewCallback(ctx context.Context, client *telegram.Bot, query *models.CallbackQuery) {
	if !b.isAdmin(query.From.ID) {
		b.answerCallback(ctx, client, query.ID, "Недостаточно прав.", true)
		return
	}

	taskID, err := uuid.Parse(strings.TrimPrefix(query.Data, adminPreviewPrefix))
	if err != nil {
		b.answerCallback(ctx, client, query.ID, "Некорректный таск.", true)
		return
	}

	tasks, err := b.services.Tasks.GetUpcoming(ctx, 100)
	if err != nil {
		b.answerCallback(ctx, client, query.ID, "Не удалось получить таск.", true)
		b.handleError(ctx, query.From.ID, err)
		return
	}
	for _, task := range tasks {
		if task.ID != taskID {
			continue
		}

		b.answerCallback(ctx, client, query.ID, "Тестовая отправка выполнена.", false)
		if err := b.sendTaskPreview(ctx, query.From.ID, task); err != nil {
			b.handleError(ctx, query.From.ID, err)
		}
		return
	}

	b.answerCallback(ctx, client, query.ID, "Таск уже открыт или отключён.", true)
}

func (b *Bot) handleAdmin(ctx context.Context, chatID, telegramUserID int64) {
	if !b.isAdmin(telegramUserID) {
		b.sendText(ctx, chatID, "Недостаточно прав.", b.mainMenu(telegramUserID))
		return
	}

	tasks, err := b.services.Tasks.GetUpcoming(ctx, 20)
	if err != nil {
		b.handleError(ctx, chatID, err)
		return
	}
	if len(tasks) == 0 {
		b.sendText(ctx, chatID, "Предстоящих тасков пока нет.", b.mainMenu(telegramUserID))
		return
	}

	b.sendText(ctx, chatID, formatUpcomingTasks(tasks), upcomingTasksKeyboard(tasks))
}

func (b *Bot) answerCallback(
	ctx context.Context,
	client *telegram.Bot,
	callbackID, text string,
	showAlert bool,
) {
	if _, err := client.AnswerCallbackQuery(ctx, &telegram.AnswerCallbackQueryParams{
		CallbackQueryID: callbackID,
		Text:            text,
		ShowAlert:       showAlert,
	}); err != nil {
		b.logError(ctx, "Failed to answer callback query", err)
	}
}

func (b *Bot) handleFlag(ctx context.Context, chatID, telegramUserID int64, flag string) {
	taskID, exists := b.getPendingTask(telegramUserID)
	if !exists {
		b.sendText(ctx, chatID, "Выберите действие в меню.", b.mainMenu(telegramUserID))
		return
	}

	result, err := b.services.Submissions.Submit(ctx, dto.SubmitTask{
		TelegramUserID: telegramUserID,
		TaskID:         taskID,
		Flag:           flag,
	})
	if err != nil {
		b.handleError(ctx, chatID, err)
		return
	}

	switch {
	case result.AlreadySolved:
		b.deletePendingTask(telegramUserID)
		b.sendText(ctx, chatID, "Этот таск уже решён вами.", b.mainMenu(telegramUserID))
	case result.Ignored && result.Correct:
		b.deletePendingTask(telegramUserID)
		b.sendText(
			ctx,
			chatID,
			"Флаг верный. Попытка администратора не сохранена и не влияет на рейтинг.",
			b.mainMenu(telegramUserID),
		)
	case result.Correct:
		b.deletePendingTask(telegramUserID)
		b.sendText(
			ctx,
			chatID,
			fmt.Sprintf(
				"Верно! Начислено очков: %d\nВсего очков: %d",
				result.PointsAwarded,
				result.TotalPoints,
			),
			b.mainMenu(telegramUserID),
		)
	default:
		b.sendText(ctx, chatID, "Неверный флаг. Попробуйте ещё раз.", nil)
	}
}

func (b *Bot) authenticate(ctx context.Context, user *models.User) (*dto.User, error) {
	return b.services.Users.Authenticate(ctx, dto.AuthenticateUser{
		TelegramUserID:   user.ID,
		TelegramUsername: optionalString(user.Username),
		FirstName:        optionalString(user.FirstName),
		LastName:         optionalString(user.LastName),
	})
}

func (b *Bot) handleError(ctx context.Context, chatID int64, err error) {
	b.logError(
		ctx,
		"Failed to handle Telegram update",
		err,
		"chat_id", chatID,
	)

	message := "Не удалось выполнить действие. Попробуйте позже."
	switch {
	case errors.Is(err, repository.ErrNotFound):
		message = "Запрошенные данные не найдены."
	case errors.Is(err, service.ErrInvalidInput):
		message = "Переданы некорректные данные."
	case errors.Is(err, service.ErrTaskUnavailable):
		message = "Этот таск сейчас недоступен."
	case errors.Is(err, service.ErrUserInactive):
		message = "Ваш аккаунт заблокирован."
	}

	b.sendText(ctx, chatID, message, b.mainMenu(chatID))
}

func optionalString(value string) *string {
	if value == "" {
		return nil
	}

	return &value
}

func (b *Bot) sendUnavailable(ctx context.Context, chatID int64) {
	now := time.Now()
	message := "Бот завершил работу."
	if now.Before(b.startsAt) {
		message = fmt.Sprintf(
			"Бот начнёт работу %s.",
			b.startsAt.Format("02.01.2006 15:04 MST"),
		)
	}

	b.sendText(ctx, chatID, message, nil)
}
