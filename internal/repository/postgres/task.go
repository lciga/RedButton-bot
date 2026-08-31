package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"RedButton-bot/internal/database/model"
	"RedButton-bot/internal/repository"
	"RedButton-bot/internal/scoring"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Структура репозитория тасков PostgreSQL
type taskRepository struct {
	db *gorm.DB
}

// Функция синхронизации тасков из YAML-файлов.
// Обновляет существующие таски и деактивирует отсутствующие.
func (r *taskRepository) Sync(ctx context.Context, tasks []model.Task) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		slugs := make([]string, 0, len(tasks))
		for _, task := range tasks {
			slugs = append(slugs, task.Slug)
		}
		if err := tx.Model(&model.Task{}).
			Where("slug NOT IN ?", slugs).
			Update("is_active", false).
			Error; err != nil {
			return err
		}

		for index := range tasks {
			if err := syncTask(tx, &tasks[index]); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("synchronize tasks: %w", err)
	}

	return nil
}

func syncTask(tx *gorm.DB, task *model.Task) error {
	files := task.Files
	task.Files = nil

	var stored model.Task
	err := tx.First(&stored, "slug = ?", task.Slug).Error
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		if err := tx.Create(task).Error; err != nil {
			return err
		}
	case err != nil:
		return err
	default:
		var correctSubmissions []model.Submission
		if err := tx.
			Where("task_id = ? AND is_correct = ?", stored.ID, true).
			Order("submitted_at ASC, id ASC").
			Find(&correctSubmissions).
			Error; err != nil {
			return err
		}
		currentPoints := scoring.Calculate(
			task.InitialPoints,
			task.MinimumPoints,
			task.Decay,
			int64(len(correctSubmissions)+1),
		)
		scoringChanged := stored.InitialPoints != task.InitialPoints ||
			stored.MinimumPoints != task.MinimumPoints ||
			stored.Decay != task.Decay
		if !scoringChanged {
			for index := range correctSubmissions {
				expectedPoints := scoring.Calculate(
					task.InitialPoints,
					task.MinimumPoints,
					task.Decay,
					int64(index+1),
				)
				if correctSubmissions[index].PointsAwarded != expectedPoints {
					scoringChanged = true
					break
				}
			}
		}

		if err := tx.Model(&stored).Updates(map[string]any{
			"title":          task.Title,
			"description":    task.Description,
			"flag_hash":      task.FlagHash,
			"initial_points": task.InitialPoints,
			"minimum_points": task.MinimumPoints,
			"current_points": currentPoints,
			"decay":          task.Decay,
			"starts_at":      task.StartsAt,
			"ends_at":        task.EndsAt,
			"is_active":      true,
		}).Error; err != nil {
			return err
		}
		if scoringChanged {
			if err := recalculateTaskScores(tx, task, correctSubmissions); err != nil {
				return err
			}
		}
		task.ID = stored.ID
	}

	if err := tx.Where("task_id = ?", task.ID).Delete(&model.TaskFile{}).Error; err != nil {
		return err
	}
	for index := range files {
		files[index].ID = uuid.Nil
		files[index].TaskID = task.ID
		if err := tx.Create(&files[index]).Error; err != nil {
			return err
		}
	}

	return nil
}

func recalculateTaskScores(tx *gorm.DB, task *model.Task, submissions []model.Submission) error {
	userIDs := make(map[uuid.UUID]struct{}, len(submissions))
	for index := range submissions {
		points := scoring.Calculate(
			task.InitialPoints,
			task.MinimumPoints,
			task.Decay,
			int64(index+1),
		)
		if err := tx.Model(&model.Submission{}).
			Where("id = ?", submissions[index].ID).
			Update("points_awarded", points).
			Error; err != nil {
			return err
		}
		userIDs[submissions[index].UserID] = struct{}{}
	}

	for userID := range userIDs {
		if err := rebuildUserRating(tx, userID); err != nil {
			return err
		}
	}

	return nil
}

func rebuildUserRating(tx *gorm.DB, userID uuid.UUID) error {
	var aggregate struct {
		TotalPoints      int
		SolvedTasksCount int
		LastSolvedAt     *time.Time
	}
	if err := tx.Model(&model.Submission{}).
		Select(`
			COALESCE(SUM(points_awarded), 0) AS total_points,
			COUNT(*) AS solved_tasks_count,
			MAX(submitted_at) AS last_solved_at
		`).
		Where("user_id = ? AND is_correct = ?", userID, true).
		Scan(&aggregate).
		Error; err != nil {
		return err
	}

	now := time.Now()
	rating := model.Rating{
		UserID:           userID,
		TotalPoints:      aggregate.TotalPoints,
		SolvedTasksCount: aggregate.SolvedTasksCount,
		LastSolvedAt:     aggregate.LastSolvedAt,
		UpdatedAt:        now,
	}
	return tx.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.Assignments(map[string]any{
			"total_points":       aggregate.TotalPoints,
			"solved_tasks_count": aggregate.SolvedTasksCount,
			"last_solved_at":     aggregate.LastSolvedAt,
			"updated_at":         now,
		}),
	}).Create(&rating).Error
}

// Функция получения таска по идентификатору.
// Возвращает таск вместе с прикреплёнными файлами.
func (r *taskRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Task, error) {
	var task model.Task
	err := r.db.WithContext(ctx).
		Preload("Files").
		First(&task, "id = ?", id).
		Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get task: %w", err)
	}

	return &task, nil
}

// Функция получения таска с блокировкой строки.
// Используется во время сдачи решения.
func (r *taskRepository) GetByIDForUpdate(ctx context.Context, id uuid.UUID) (*model.Task, error) {
	var task model.Task
	err := r.db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&task, "id = ?", id).
		Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get task for update: %w", err)
	}

	return &task, nil
}

// Функция получения новой таски для пользователя.
// Возвращает последнюю доступную и ещё не решённую таску.
func (r *taskRepository) GetNewestAvailableForUser(
	ctx context.Context,
	telegramUserID int64,
	now time.Time,
) (*model.Task, error) {
	var task model.Task
	err := r.db.WithContext(ctx).
		Preload("Files").
		Where("tasks.is_active = ?", true).
		Where("tasks.starts_at <= ?", now).
		Where("tasks.ends_at IS NULL OR tasks.ends_at > ?", now).
		Where(`NOT EXISTS (
			SELECT 1
			FROM submissions
			JOIN users ON users.id = submissions.user_id
			WHERE submissions.task_id = tasks.id
			  AND submissions.is_correct = TRUE
			  AND users.telegram_user_id = ?
		)`, telegramUserID).
		Order("tasks.starts_at DESC, tasks.id DESC").
		First(&task).
		Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get newest task for user: %w", err)
	}

	return &task, nil
}

// Функция получения доступных тасков.
// Возвращает активные и открытые на указанное время таски.
func (r *taskRepository) ListAvailable(ctx context.Context, now time.Time) ([]model.Task, error) {
	var tasks []model.Task
	err := r.db.WithContext(ctx).
		Preload("Files").
		Where("is_active = ?", true).
		Where("starts_at <= ?", now).
		Where("ends_at IS NULL OR ends_at > ?", now).
		Order("starts_at ASC, id ASC").
		Find(&tasks).
		Error
	if err != nil {
		return nil, fmt.Errorf("list available tasks: %w", err)
	}

	return tasks, nil
}

// Функция получения решённых и доступных сейчас тасков для профиля.
func (r *taskRepository) ListForProfile(
	ctx context.Context,
	userID uuid.UUID,
	now time.Time,
) ([]repository.ProfileTask, error) {
	var tasks []repository.ProfileTask
	solvedExpression := `EXISTS (
		SELECT 1 FROM submissions
		WHERE submissions.task_id = tasks.id
		  AND submissions.user_id = ?
		  AND submissions.is_correct = TRUE
	)`
	err := r.db.WithContext(ctx).
		Model(&model.Task{}).
		Select("tasks.id, tasks.title, "+solvedExpression+" AS solved", userID).
		Where(
			"("+solvedExpression+") OR (tasks.is_active = ? AND tasks.starts_at <= ? AND (tasks.ends_at IS NULL OR tasks.ends_at > ?))",
			userID,
			true,
			now,
			now,
		).
		Order("solved DESC, tasks.starts_at ASC, tasks.id ASC").
		Scan(&tasks).
		Error
	if err != nil {
		return nil, fmt.Errorf("list profile tasks: %w", err)
	}
	return tasks, nil
}

// Функция получения времени открытия ближайшего таска.
// Возвращает nil при отсутствии предстоящих тасков.
func (r *taskRepository) GetNextStartsAt(ctx context.Context, now time.Time) (*time.Time, error) {
	var result struct {
		StartsAt *time.Time
	}
	err := r.db.WithContext(ctx).
		Model(&model.Task{}).
		Select("MIN(starts_at) AS starts_at").
		Where("is_active = ?", true).
		Where("starts_at > ?", now).
		Scan(&result).
		Error
	if err != nil {
		return nil, fmt.Errorf("get next task start time: %w", err)
	}

	return result.StartsAt, nil
}

// Функция получения предстоящих тасков.
// Возвращает активные таски до наступления времени открытия.
func (r *taskRepository) ListUpcoming(ctx context.Context, now time.Time, limit int) ([]model.Task, error) {
	var tasks []model.Task
	err := r.db.WithContext(ctx).
		Preload("Files").
		Where("is_active = ?", true).
		Where("starts_at > ?", now).
		Order("starts_at ASC, id ASC").
		Limit(limit).
		Find(&tasks).
		Error
	if err != nil {
		return nil, fmt.Errorf("list upcoming tasks: %w", err)
	}

	return tasks, nil
}

// Функция обновления текущей стоимости таска.
// Возвращает ошибку обновления.
func (r *taskRepository) UpdateCurrentPoints(ctx context.Context, id uuid.UUID, points int) error {
	result := r.db.WithContext(ctx).
		Model(&model.Task{}).
		Where("id = ?", id).
		Update("current_points", points)
	if result.Error != nil {
		return fmt.Errorf("update task points: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return repository.ErrNotFound
	}

	return nil
}
