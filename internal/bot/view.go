package bot

import (
	"context"
	"fmt"
	"os"
	"strings"

	"RedButton-bot/internal/dto"
	telegram "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func (b *Bot) mainMenu(telegramUserID int64) *models.ReplyKeyboardMarkup {
	keyboard := [][]models.KeyboardButton{
		{
			{Text: buttonNewTask},
			{Text: buttonRating},
		},
	}
	if b.isAdmin(telegramUserID) {
		keyboard = append(keyboard, []models.KeyboardButton{{Text: buttonAdmin}})
	}

	return &models.ReplyKeyboardMarkup{
		Keyboard:       keyboard,
		IsPersistent:   true,
		ResizeKeyboard: true,
	}
}

func upcomingTasksKeyboard(tasks []dto.Task) *models.InlineKeyboardMarkup {
	keyboard := make([][]models.InlineKeyboardButton, 0, len(tasks))
	for _, task := range tasks {
		keyboard = append(keyboard, []models.InlineKeyboardButton{
			{
				Text:         "🧪 " + task.Title,
				CallbackData: adminPreviewPrefix + task.ID.String(),
			},
		})
	}

	return &models.InlineKeyboardMarkup{InlineKeyboard: keyboard}
}

func solveKeyboard(taskID string) *models.InlineKeyboardMarkup {
	return &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "🚩 Отправить флаг", CallbackData: solvePrefix + taskID},
			},
		},
	}
}

func ratingKeyboard(page dto.RatingPage) models.ReplyMarkup {
	buttons := make([]models.InlineKeyboardButton, 0, 2)
	if page.Page > 1 {
		buttons = append(buttons, models.InlineKeyboardButton{
			Text:         "⬅️ Назад",
			CallbackData: ratingPrefix + fmt.Sprint(page.Page-1),
		})
	}
	if page.Page < page.TotalPages {
		buttons = append(buttons, models.InlineKeyboardButton{
			Text:         "Вперёд ➡️",
			CallbackData: ratingPrefix + fmt.Sprint(page.Page+1),
		})
	}
	if len(buttons) == 0 {
		return nil
	}

	return &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{buttons},
	}
}

func (b *Bot) sendText(ctx context.Context, chatID int64, text string, markup models.ReplyMarkup) {
	if _, err := b.client.SendMessage(ctx, &telegram.SendMessageParams{
		ChatID:      chatID,
		Text:        text,
		ReplyMarkup: markup,
	}); err != nil {
		b.logError(ctx, "Failed to send Telegram message", err, "chat_id", chatID)
	}
}

func (b *Bot) sendTask(ctx context.Context, chatID int64, task dto.Task, notification bool) error {
	title := "Доступна таска"
	if notification {
		title = "🔔 Открылась новая таска"
	}

	return b.sendTaskContent(ctx, chatID, task, title, solveKeyboard(task.ID.String()))
}

func (b *Bot) sendTaskPreview(ctx context.Context, chatID int64, task dto.Task) error {
	return b.sendTaskContent(ctx, chatID, task, "🧪 Тестовая отправка таска", nil)
}

func (b *Bot) sendTaskContent(
	ctx context.Context,
	chatID int64,
	task dto.Task,
	title string,
	markup models.ReplyMarkup,
) error {
	text := fmt.Sprintf(
		"%s\n\n%s\n\n%s\n\nСтоимость: %d очков",
		title,
		task.Title,
		task.Description,
		task.CurrentPoints,
	)
	if _, err := b.client.SendMessage(ctx, &telegram.SendMessageParams{
		ChatID:      chatID,
		Text:        text,
		ReplyMarkup: markup,
	}); err != nil {
		return fmt.Errorf("send task: %w", err)
	}

	for _, file := range task.Files {
		if err := b.sendTaskFile(ctx, chatID, file); err != nil {
			return err
		}
	}

	return nil
}

func formatUpcomingTasks(tasks []dto.Task) string {
	var result strings.Builder
	result.WriteString("⚙️ Предстоящие таски\n\n")
	for index, task := range tasks {
		fmt.Fprintf(
			&result,
			"%d. %s\nОткрытие: %s\nСтоимость: %d очков\n\n",
			index+1,
			task.Title,
			task.StartsAt.Format("02.01.2006 15:04 MST"),
			task.CurrentPoints,
		)
	}
	result.WriteString("Выберите таск для тестовой отправки:")

	return result.String()
}

func (b *Bot) sendTaskFile(ctx context.Context, chatID int64, file dto.TaskFile) error {
	var input models.InputFile
	if file.TelegramFileID != nil && *file.TelegramFileID != "" {
		input = &models.InputFileString{Data: *file.TelegramFileID}
	} else {
		storedFile, err := os.Open(file.StoragePath)
		if err != nil {
			return fmt.Errorf("open task file %q: %w", file.FileName, err)
		}
		defer storedFile.Close()
		input = &models.InputFileUpload{
			Filename: file.FileName,
			Data:     storedFile,
		}
	}

	if _, err := b.client.SendDocument(ctx, &telegram.SendDocumentParams{
		ChatID:   chatID,
		Document: input,
		Caption:  file.FileName,
	}); err != nil {
		return fmt.Errorf("send task file %q: %w", file.FileName, err)
	}

	return nil
}

func formatRating(page dto.RatingPage) string {
	if len(page.Items) == 0 {
		return "Рейтинг пока пуст."
	}

	var result strings.Builder
	result.WriteString("🏆 Рейтинг\n\n")
	for _, item := range page.Items {
		fmt.Fprintf(
			&result,
			"%d. %s — %d очков (%d тасок)\n",
			item.Position,
			formatUserName(item),
			item.TotalPoints,
			item.SolvedTasksCount,
		)
	}
	fmt.Fprintf(&result, "\nСтраница %d из %d", page.Page, page.TotalPages)

	return strings.TrimSpace(result.String())
}

func formatUserName(user dto.Rating) string {
	if user.TelegramUsername != nil && *user.TelegramUsername != "" {
		return "@" + *user.TelegramUsername
	}

	parts := make([]string, 0, 2)
	if user.FirstName != nil && *user.FirstName != "" {
		parts = append(parts, *user.FirstName)
	}
	if user.LastName != nil && *user.LastName != "" {
		parts = append(parts, *user.LastName)
	}
	if len(parts) != 0 {
		return strings.Join(parts, " ")
	}

	return fmt.Sprintf("Пользователь %d", user.TelegramUserID)
}
