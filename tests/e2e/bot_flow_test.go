//go:build e2e

package e2e_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"RedButton-bot/internal/dto"
	postgresrepository "RedButton-bot/internal/repository/postgres"
	"RedButton-bot/internal/service"
	"RedButton-bot/internal/taskconfig"
	"RedButton-bot/internal/testsupport"
)

func TestCompleteParticipantFlow(t *testing.T) {
	db := testsupport.PostgreSQL(t)
	store := postgresrepository.New(db)
	repositories := store.Repositories()
	services := service.New(repositories, store, nil, nil, time.Nanosecond)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)
	location := time.FixedZone("UTC+5", 5*60*60)
	startsAt := now.Add(-time.Hour).In(location)

	directory := t.TempDir()
	yaml := fmt.Sprintf("title: E2E task\ndescription: Complete flow\nflag: redbutton{e2e}\nmaximum_points: 100\nminimum_points: 10\ndecay: 5\nstarts_at: %q\n", startsAt.Format("2006-01-02T15:04:05"))
	if err := os.WriteFile(filepath.Join(directory, "e2e.yaml"), []byte(yaml), 0o600); err != nil {
		t.Fatal(err)
	}
	definitions, err := taskconfig.Load(directory, location)
	if err != nil {
		t.Fatal(err)
	}
	if err := services.Tasks.Sync(ctx, definitions, 3*time.Hour, now.Add(-2*time.Hour), now.Add(2*time.Hour)); err != nil {
		t.Fatal(err)
	}

	first, err := services.Users.Authenticate(ctx, dto.AuthenticateUser{TelegramUserID: 1001})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := services.Users.Authenticate(ctx, dto.AuthenticateUser{TelegramUserID: 1002}); err != nil {
		t.Fatal(err)
	}
	task, err := services.Tasks.GetNewestForUser(ctx, 1001)
	if err != nil {
		t.Fatal(err)
	}

	wrong, err := services.Submissions.Submit(ctx, dto.SubmitTask{TelegramUserID: 1001, TaskID: task.ID, Flag: "wrong"})
	if err != nil || wrong.Correct || wrong.PointsAwarded != 0 {
		t.Fatalf("wrong result=%#v error=%v", wrong, err)
	}
	correct, err := services.Submissions.Submit(ctx, dto.SubmitTask{TelegramUserID: 1001, TaskID: task.ID, Flag: "redbutton{e2e}"})
	if err != nil || !correct.Correct || correct.PointsAwarded != 100 || correct.CurrentPoints != 97 {
		t.Fatalf("correct result=%#v error=%v", correct, err)
	}
	repeated, err := services.Submissions.Submit(ctx, dto.SubmitTask{TelegramUserID: 1001, TaskID: task.ID, Flag: "redbutton{e2e}"})
	if err != nil || !repeated.AlreadySolved {
		t.Fatalf("repeat result=%#v error=%v", repeated, err)
	}
	second, err := services.Submissions.Submit(ctx, dto.SubmitTask{TelegramUserID: 1002, TaskID: task.ID, Flag: "redbutton{e2e}"})
	if err != nil || second.PointsAwarded != 97 {
		t.Fatalf("second result=%#v error=%v", second, err)
	}
	profile, err := services.Profiles.Get(ctx, 1001)
	if err != nil || profile.TotalPoints != 100 || profile.Position != 1 || len(profile.Tasks) != 1 || !profile.Tasks[0].Solved {
		t.Fatalf("profile=%#v error=%v", profile, err)
	}

	leaderboard, err := services.Ratings.GetLeaderboardPage(ctx, 1, 10)
	if err != nil || leaderboard.TotalItems != 2 || len(leaderboard.Items) != 2 || leaderboard.Items[0].TotalPoints != 100 {
		t.Fatalf("leaderboard=%#v error=%v", leaderboard, err)
	}
	if _, err := services.Ratings.GetByUserID(ctx, first.ID); err != nil {
		t.Fatal(err)
	}

	pending, err := services.Notifications.GetPending(ctx, 10)
	if err != nil || len(pending) != 2 {
		t.Fatalf("pending=%#v error=%v", pending, err)
	}
	if err := services.Notifications.MarkSent(ctx, pending[0].UserID, pending[0].Task.ID); err != nil {
		t.Fatal(err)
	}
	pending, err = services.Notifications.GetPending(ctx, 10)
	if err != nil || len(pending) != 1 {
		t.Fatalf("pending after mark=%#v error=%v", pending, err)
	}
}
