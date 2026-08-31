package dto

import "github.com/google/uuid"

// Структура данных задачи в профиле пользователя.
type ProfileTask struct {
	ID     uuid.UUID
	Title  string
	Solved bool
}

// Структура данных профиля пользователя.
type Profile struct {
	TotalPoints int
	Position    int
	Tasks       []ProfileTask
}
