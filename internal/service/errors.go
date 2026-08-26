package service

import "errors"

var (
	ErrInvalidInput    = errors.New("invalid input")
	ErrUserInactive    = errors.New("user is inactive")
	ErrTaskUnavailable = errors.New("task is unavailable")
)
