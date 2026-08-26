package service

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"strings"
	"time"

	"RedButton-bot/internal/database/model"
	"RedButton-bot/internal/dto"
	"RedButton-bot/internal/repository"
	"RedButton-bot/internal/scoring"
	"github.com/google/uuid"
)

// Интерфейс проверки флага
type FlagVerifier interface {
	Verify(flag, hash string) bool
}

// Структура проверки SHA-256 хэша флага
type SHA256FlagVerifier struct{}

// Функция проверки флага по SHA-256 хэшу.
// Возвращает true при совпадении хэшей.
func (SHA256FlagVerifier) Verify(flag, hash string) bool {
	expected, err := hex.DecodeString(hash)
	if err != nil || len(expected) != sha256.Size {
		return false
	}

	actual := sha256.Sum256([]byte(flag))
	return subtle.ConstantTimeCompare(actual[:], expected) == 1
}

// Функция вычисления SHA-256 хэша флага.
// Возвращает строку для сохранения в базе данных.
func HashFlag(flag string) string {
	hash := sha256.Sum256([]byte(flag))
	return hex.EncodeToString(hash[:])
}

// Структура сервиса попыток сдачи
type SubmissionService struct {
	transactor          repository.Transactor
	verifier            FlagVerifier
	now                 func() time.Time
	excludedTelegramIDs map[int64]struct{}
}

// Функция создания сервиса попыток сдачи.
// Возвращает указатель на экземпляр SubmissionService.
func NewSubmissionService(
	transactor repository.Transactor,
	verifier FlagVerifier,
	excludedTelegramIDs map[int64]struct{},
) *SubmissionService {
	if verifier == nil {
		verifier = SHA256FlagVerifier{}
	}

	excluded := make(map[int64]struct{}, len(excludedTelegramIDs))
	for telegramID := range excludedTelegramIDs {
		excluded[telegramID] = struct{}{}
	}

	return &SubmissionService{
		transactor:          transactor,
		verifier:            verifier,
		now:                 time.Now,
		excludedTelegramIDs: excluded,
	}
}

// Функция проверки и сохранения решения таска.
// Возвращает результат попытки и начисленные очки.
func (s *SubmissionService) Submit(ctx context.Context, input dto.SubmitTask) (*dto.SubmissionResult, error) {
	if input.TelegramUserID <= 0 || input.TaskID == uuid.Nil || strings.TrimSpace(input.Flag) == "" {
		return nil, ErrInvalidInput
	}

	result := dto.SubmissionResult{TaskID: input.TaskID}
	err := s.transactor.WithinTransaction(ctx, func(repositories repository.Repositories) error {
		user, err := repositories.Users.GetByTelegramID(ctx, input.TelegramUserID)
		if err != nil {
			return err
		}
		if !user.IsActive {
			return ErrUserInactive
		}

		task, err := repositories.Tasks.GetByIDForUpdate(ctx, input.TaskID)
		if err != nil {
			return err
		}

		now := s.now()
		if !task.IsActive || task.StartsAt.After(now) || task.EndsAt != nil && !task.EndsAt.After(now) {
			return ErrTaskUnavailable
		}

		correct := s.verifier.Verify(input.Flag, task.FlagHash)
		if _, excluded := s.excludedTelegramIDs[input.TelegramUserID]; excluded {
			result.Correct = correct
			result.Ignored = true
			result.CurrentPoints = task.CurrentPoints
			return nil
		}

		solved, err := repositories.Submissions.HasCorrect(ctx, user.ID, task.ID)
		if err != nil {
			return err
		}
		if solved {
			result.AlreadySolved = true
			result.CurrentPoints = task.CurrentPoints
			return nil
		}

		correctCount, err := repositories.Submissions.CountCorrect(ctx, task.ID)
		if err != nil {
			return err
		}
		points := task.CurrentPoints
		if correct {
			points = scoring.Calculate(
				task.InitialPoints,
				task.MinimumPoints,
				task.Decay,
				correctCount+1,
			)
		}

		submission := model.Submission{
			UserID:      user.ID,
			TaskID:      task.ID,
			IsCorrect:   correct,
			SubmittedAt: now,
		}
		if correct {
			submission.PointsAwarded = points
		}
		if err := repositories.Submissions.Create(ctx, &submission); err != nil {
			return err
		}

		result.Correct = correct
		result.CurrentPoints = task.CurrentPoints
		if !correct {
			return nil
		}

		if err := repositories.Ratings.AddSolution(ctx, user.ID, points, now); err != nil {
			return err
		}
		result.PointsAwarded = points

		nextPoints := calculateNextPoints(task, correctCount+2)
		if nextPoints != task.CurrentPoints {
			if err := repositories.Tasks.UpdateCurrentPoints(ctx, task.ID, nextPoints); err != nil {
				return err
			}
			result.CurrentPoints = nextPoints
		}

		rating, err := repositories.Ratings.GetByUserID(ctx, user.ID)
		if err != nil {
			return err
		}
		result.TotalPoints = rating.TotalPoints

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func calculateNextPoints(task *model.Task, correctCount int64) int {
	return scoring.Calculate(task.InitialPoints, task.MinimumPoints, task.Decay, correctCount)
}
