package service

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrInvalidInput    = errors.New("invalid input")
	ErrUserInactive    = errors.New("user is inactive")
	ErrTaskUnavailable = errors.New("task is unavailable")
	ErrRateLimited     = errors.New("submission rate limit exceeded")
)

// Ошибка ограничения частоты сдачи флагов.
type RateLimitError struct {
	RetryAfter time.Duration
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("%s: retry after %s", ErrRateLimited, e.RetryAfter)
}

func (e *RateLimitError) Unwrap() error {
	return ErrRateLimited
}
