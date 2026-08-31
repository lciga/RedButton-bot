package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"RedButton-bot/internal/database/model"
	"RedButton-bot/internal/dto"
	"RedButton-bot/internal/repository"
	"github.com/google/uuid"
)

func TestSHA256FlagVerifier(t *testing.T) {
	verifier := SHA256FlagVerifier{}
	hash := HashFlag("redbutton{correct}")

	if !verifier.Verify("redbutton{correct}", hash) {
		t.Error("Verify() = false, want true")
	}
	if verifier.Verify("redbutton{wrong}", hash) {
		t.Error("Verify() = true, want false")
	}
}

func TestSubmitCorrectSolution(t *testing.T) {
	userID, taskID := uuid.New(), uuid.New()
	now := time.Now()
	created := false
	updatedPoints := 0
	repositories := repository.Repositories{
		Users: userRepositoryStub{get: func(context.Context, int64) (*model.User, error) {
			return &model.User{ID: userID, TelegramUserID: 42, IsActive: true}, nil
		}},
		Tasks: taskRepositoryStub{
			getForLock: func(context.Context, uuid.UUID) (*model.Task, error) {
				endsAt := now.Add(time.Hour)
				return &model.Task{ID: taskID, FlagHash: HashFlag("correct"), InitialPoints: 100, MinimumPoints: 10, CurrentPoints: 100, Decay: 5, StartsAt: now.Add(-time.Hour), EndsAt: &endsAt, IsActive: true}, nil
			},
			update: func(_ context.Context, _ uuid.UUID, points int) error { updatedPoints = points; return nil },
		},
		Submissions: submissionRepositoryStub{
			has:   func(context.Context, uuid.UUID, uuid.UUID) (bool, error) { return false, nil },
			count: func(context.Context, uuid.UUID) (int64, error) { return 0, nil },
			create: func(_ context.Context, submission *model.Submission) error {
				created = submission.IsCorrect && submission.PointsAwarded == 100
				return nil
			},
		},
		Ratings: ratingRepositoryStub{
			add: func(_ context.Context, _ uuid.UUID, points int, _ time.Time) error {
				if points != 100 {
					t.Fatalf("points = %d, want 100", points)
				}
				return nil
			},
			get: func(context.Context, uuid.UUID) (*model.Rating, error) { return &model.Rating{TotalPoints: 100}, nil },
		},
	}
	service := NewSubmissionService(transactorStub{repositories: repositories}, nil, nil, 5*time.Second)
	service.now = func() time.Time { return now }
	result, err := service.Submit(context.Background(), dto.SubmitTask{TelegramUserID: 42, TaskID: taskID, Flag: "correct"})
	if err != nil {
		t.Fatal(err)
	}
	if !created || !result.Correct || result.PointsAwarded != 100 || result.TotalPoints != 100 || updatedPoints != 97 {
		t.Fatalf("created=%v updated=%d result=%#v", created, updatedPoints, result)
	}
}

func TestSubmitValidationUnavailableAndIgnored(t *testing.T) {
	service := NewSubmissionService(transactorStub{}, nil, nil, 5*time.Second)
	if _, err := service.Submit(context.Background(), dto.SubmitTask{}); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("error = %v", err)
	}

	userID, taskID := uuid.New(), uuid.New()
	repositories := repository.Repositories{
		Users: userRepositoryStub{get: func(context.Context, int64) (*model.User, error) {
			return &model.User{ID: userID, IsActive: false}, nil
		}},
	}
	service = NewSubmissionService(transactorStub{repositories: repositories}, nil, nil, 5*time.Second)
	if _, err := service.Submit(context.Background(), dto.SubmitTask{TelegramUserID: 1, TaskID: taskID, Flag: "x"}); !errors.Is(err, ErrUserInactive) {
		t.Fatalf("error = %v", err)
	}

	now := time.Now()
	repositories.Users = userRepositoryStub{get: func(context.Context, int64) (*model.User, error) { return &model.User{ID: userID, IsActive: true}, nil }}
	repositories.Tasks = taskRepositoryStub{getForLock: func(context.Context, uuid.UUID) (*model.Task, error) {
		return &model.Task{ID: taskID, StartsAt: now.Add(time.Hour), IsActive: true}, nil
	}}
	service = NewSubmissionService(transactorStub{repositories: repositories}, nil, nil, 5*time.Second)
	service.now = func() time.Time { return now }
	if _, err := service.Submit(context.Background(), dto.SubmitTask{TelegramUserID: 1, TaskID: taskID, Flag: "x"}); !errors.Is(err, ErrTaskUnavailable) {
		t.Fatalf("error = %v", err)
	}

	repositories.Tasks = taskRepositoryStub{getForLock: func(context.Context, uuid.UUID) (*model.Task, error) {
		return &model.Task{ID: taskID, FlagHash: HashFlag("correct"), StartsAt: now.Add(-time.Hour), IsActive: true}, nil
	}}
	service = NewSubmissionService(transactorStub{repositories: repositories}, nil, map[int64]struct{}{1: {}}, 5*time.Second)
	service.now = func() time.Time { return now }
	result, err := service.Submit(context.Background(), dto.SubmitTask{TelegramUserID: 1, TaskID: taskID, Flag: "correct"})
	if err != nil || !result.Ignored || !result.Correct {
		t.Fatalf("error=%v result=%#v", err, result)
	}
}

func TestSubmitRateLimit(t *testing.T) {
	userID, taskID := uuid.New(), uuid.New()
	now := time.Now()
	lastAttempt := now.Add(-2 * time.Second)
	repositories := repository.Repositories{
		Users: userRepositoryStub{get: func(context.Context, int64) (*model.User, error) {
			return &model.User{ID: userID, IsActive: true}, nil
		}},
		Tasks: taskRepositoryStub{getForLock: func(context.Context, uuid.UUID) (*model.Task, error) {
			return &model.Task{ID: taskID, StartsAt: now.Add(-time.Hour), IsActive: true}, nil
		}},
		Submissions: submissionRepositoryStub{
			has:  func(context.Context, uuid.UUID, uuid.UUID) (bool, error) { return false, nil },
			last: func(context.Context, uuid.UUID, uuid.UUID) (*time.Time, error) { return &lastAttempt, nil },
		},
	}
	service := NewSubmissionService(transactorStub{repositories: repositories}, nil, nil, 5*time.Second)
	service.now = func() time.Time { return now }
	_, err := service.Submit(context.Background(), dto.SubmitTask{TelegramUserID: 42, TaskID: taskID, Flag: "guess"})
	var rateLimitError *RateLimitError
	if !errors.As(err, &rateLimitError) || !errors.Is(err, ErrRateLimited) {
		t.Fatalf("error = %v, want RateLimitError", err)
	}
	if rateLimitError.RetryAfter != 3*time.Second {
		t.Fatalf("retry after = %v, want 3s", rateLimitError.RetryAfter)
	}
}

func TestCalculateNextPoints(t *testing.T) {
	task := &model.Task{
		InitialPoints: 100,
		MinimumPoints: 10,
		CurrentPoints: 100,
		Decay:         5,
	}

	tests := []struct {
		name         string
		correctCount int64
		want         int
	}{
		{name: "first solve", correctCount: 1, want: 100},
		{name: "second solve", correctCount: 2, want: 97},
		{name: "decay reached", correctCount: 6, want: 10},
		{name: "minimum points", correctCount: 200, want: 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := calculateNextPoints(task, tt.correctCount); got != tt.want {
				t.Errorf("calculateNextPoints() = %d, want %d", got, tt.want)
			}
		})
	}
}
