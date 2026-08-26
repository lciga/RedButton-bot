package bot

import (
	"strings"
	"testing"
	"time"

	"RedButton-bot/internal/dto"
	"github.com/google/uuid"
)

func TestFormatRating(t *testing.T) {
	username := "alice"
	page := dto.RatingPage{Page: 1, TotalPages: 2, Items: []dto.Rating{{Position: 1, TelegramUsername: &username, TotalPoints: 100, SolvedTasksCount: 1}}}
	got := formatRating(page)
	for _, want := range []string{"@alice", "100 очков", "Страница 1 из 2"} {
		if !strings.Contains(got, want) {
			t.Errorf("formatRating() = %q, missing %q", got, want)
		}
	}
	if formatRating(dto.RatingPage{}) != "Рейтинг пока пуст." {
		t.Error("empty rating text is incorrect")
	}
}

func TestKeyboardsAndBotState(t *testing.T) {
	id := uuid.New()
	if keyboard := solveKeyboard(id.String()); keyboard.InlineKeyboard[0][0].CallbackData != solvePrefix+id.String() {
		t.Fatal("solve callback is incorrect")
	}
	if ratingKeyboard(dto.RatingPage{Page: 1, TotalPages: 1}) != nil {
		t.Fatal("single rating page must not have navigation")
	}

	now := time.Now()
	bot := &Bot{adminTelegramIDs: map[int64]struct{}{7: {}}, startsAt: now.Add(-time.Minute), endsAt: now.Add(time.Minute)}
	if !bot.isAdmin(7) || bot.isAdmin(8) || !bot.isAvailable(now) {
		t.Fatal("bot access state is incorrect")
	}
	bot.setPendingTask(7, id)
	if got, ok := bot.getPendingTask(7); !ok || got != id {
		t.Fatal("pending task was not stored")
	}
	bot.deletePendingTask(7)
	if _, ok := bot.getPendingTask(7); ok {
		t.Fatal("pending task was not deleted")
	}
	if len(bot.mainMenu(7).Keyboard) != 2 || len(bot.mainMenu(8).Keyboard) != 1 {
		t.Fatal("admin menu visibility is incorrect")
	}
}
