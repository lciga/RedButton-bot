package service

import (
	"context"
	"errors"
	"time"

	"RedButton-bot/internal/database/model"
	"RedButton-bot/internal/repository"
	"github.com/google/uuid"
)

var errUnexpectedCall = errors.New("unexpected repository call")

type userRepositoryStub struct {
	upsert func(context.Context, *model.User) error
	get    func(context.Context, int64) (*model.User, error)
}

func (s userRepositoryStub) Upsert(ctx context.Context, user *model.User) error {
	if s.upsert == nil {
		return errUnexpectedCall
	}
	return s.upsert(ctx, user)
}
func (s userRepositoryStub) GetByTelegramID(ctx context.Context, id int64) (*model.User, error) {
	if s.get == nil {
		return nil, errUnexpectedCall
	}
	return s.get(ctx, id)
}

type taskRepositoryStub struct {
	sync       func(context.Context, []model.Task) error
	get        func(context.Context, uuid.UUID) (*model.Task, error)
	getForLock func(context.Context, uuid.UUID) (*model.Task, error)
	newest     func(context.Context, int64, time.Time) (*model.Task, error)
	next       func(context.Context, time.Time) (*time.Time, error)
	upcoming   func(context.Context, time.Time, int) ([]model.Task, error)
	available  func(context.Context, time.Time) ([]model.Task, error)
	update     func(context.Context, uuid.UUID, int) error
}

func (s taskRepositoryStub) Sync(ctx context.Context, tasks []model.Task) error {
	if s.sync == nil {
		return errUnexpectedCall
	}
	return s.sync(ctx, tasks)
}
func (s taskRepositoryStub) GetByID(ctx context.Context, id uuid.UUID) (*model.Task, error) {
	if s.get == nil {
		return nil, errUnexpectedCall
	}
	return s.get(ctx, id)
}
func (s taskRepositoryStub) GetByIDForUpdate(ctx context.Context, id uuid.UUID) (*model.Task, error) {
	if s.getForLock == nil {
		return nil, errUnexpectedCall
	}
	return s.getForLock(ctx, id)
}
func (s taskRepositoryStub) GetNewestAvailableForUser(ctx context.Context, id int64, now time.Time) (*model.Task, error) {
	if s.newest == nil {
		return nil, errUnexpectedCall
	}
	return s.newest(ctx, id, now)
}
func (s taskRepositoryStub) GetNextStartsAt(ctx context.Context, now time.Time) (*time.Time, error) {
	if s.next == nil {
		return nil, errUnexpectedCall
	}
	return s.next(ctx, now)
}
func (s taskRepositoryStub) ListUpcoming(ctx context.Context, now time.Time, limit int) ([]model.Task, error) {
	if s.upcoming == nil {
		return nil, errUnexpectedCall
	}
	return s.upcoming(ctx, now, limit)
}
func (s taskRepositoryStub) ListAvailable(ctx context.Context, now time.Time) ([]model.Task, error) {
	if s.available == nil {
		return nil, errUnexpectedCall
	}
	return s.available(ctx, now)
}
func (s taskRepositoryStub) UpdateCurrentPoints(ctx context.Context, id uuid.UUID, points int) error {
	if s.update == nil {
		return errUnexpectedCall
	}
	return s.update(ctx, id, points)
}

type submissionRepositoryStub struct {
	create func(context.Context, *model.Submission) error
	has    func(context.Context, uuid.UUID, uuid.UUID) (bool, error)
	count  func(context.Context, uuid.UUID) (int64, error)
}

func (s submissionRepositoryStub) Create(ctx context.Context, value *model.Submission) error {
	if s.create == nil {
		return errUnexpectedCall
	}
	return s.create(ctx, value)
}
func (s submissionRepositoryStub) HasCorrect(ctx context.Context, userID, taskID uuid.UUID) (bool, error) {
	if s.has == nil {
		return false, errUnexpectedCall
	}
	return s.has(ctx, userID, taskID)
}
func (s submissionRepositoryStub) CountCorrect(ctx context.Context, taskID uuid.UUID) (int64, error) {
	if s.count == nil {
		return 0, errUnexpectedCall
	}
	return s.count(ctx, taskID)
}

type ratingRepositoryStub struct {
	add   func(context.Context, uuid.UUID, int, time.Time) error
	get   func(context.Context, uuid.UUID) (*model.Rating, error)
	list  func(context.Context, int, int, []int64) ([]model.Rating, error)
	count func(context.Context, []int64) (int64, error)
}

func (s ratingRepositoryStub) AddSolution(ctx context.Context, id uuid.UUID, points int, at time.Time) error {
	if s.add == nil {
		return errUnexpectedCall
	}
	return s.add(ctx, id, points, at)
}
func (s ratingRepositoryStub) GetByUserID(ctx context.Context, id uuid.UUID) (*model.Rating, error) {
	if s.get == nil {
		return nil, errUnexpectedCall
	}
	return s.get(ctx, id)
}
func (s ratingRepositoryStub) List(ctx context.Context, limit, offset int, excluded []int64) ([]model.Rating, error) {
	if s.list == nil {
		return nil, errUnexpectedCall
	}
	return s.list(ctx, limit, offset, excluded)
}
func (s ratingRepositoryStub) Count(ctx context.Context, excluded []int64) (int64, error) {
	if s.count == nil {
		return 0, errUnexpectedCall
	}
	return s.count(ctx, excluded)
}

type notificationRepositoryStub struct {
	list func(context.Context, time.Time, int) ([]repository.PendingTaskNotification, error)
	mark func(context.Context, uuid.UUID, uuid.UUID, time.Time) error
}

func (s notificationRepositoryStub) ListPending(ctx context.Context, now time.Time, limit int) ([]repository.PendingTaskNotification, error) {
	if s.list == nil {
		return nil, errUnexpectedCall
	}
	return s.list(ctx, now, limit)
}
func (s notificationRepositoryStub) MarkSent(ctx context.Context, userID, taskID uuid.UUID, at time.Time) error {
	if s.mark == nil {
		return errUnexpectedCall
	}
	return s.mark(ctx, userID, taskID, at)
}

type transactorStub struct {
	repositories repository.Repositories
	err          error
}

func (s transactorStub) WithinTransaction(ctx context.Context, fn func(repository.Repositories) error) error {
	if s.err != nil {
		return s.err
	}
	return fn(s.repositories)
}
