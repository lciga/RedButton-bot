package service

import "errors"

var (
	ErrInvalidInput    = errors.New("переданы некорректные данные")
	ErrUserInactive    = errors.New("пользователь заблокирован")
	ErrTaskUnavailable = errors.New("таск недоступен для решения")
)
