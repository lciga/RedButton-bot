package service

import (
	"RedButton-bot/internal/database/model"
	"RedButton-bot/internal/dto"
)

func mapUser(user *model.User) dto.User {
	return dto.User{
		ID:               user.ID,
		TelegramUserID:   user.TelegramUserID,
		TelegramUsername: user.TelegramUsername,
		FirstName:        user.FirstName,
		LastName:         user.LastName,
		IsActive:         user.IsActive,
		LastSeenAt:       user.LastSeenAt,
	}
}

func mapTask(task *model.Task) dto.Task {
	files := make([]dto.TaskFile, 0, len(task.Files))
	for _, file := range task.Files {
		files = append(files, dto.TaskFile{
			ID:             file.ID,
			FileName:       file.FileName,
			StoragePath:    file.StoragePath,
			TelegramFileID: file.TelegramFileID,
			MIMEType:       file.MIMEType,
			FileSize:       file.FileSize,
		})
	}

	return dto.Task{
		ID:            task.ID,
		Slug:          task.Slug,
		Title:         task.Title,
		Description:   task.Description,
		CurrentPoints: task.CurrentPoints,
		StartsAt:      task.StartsAt,
		EndsAt:        task.EndsAt,
		Files:         files,
	}
}

func mapRating(rating *model.Rating, position int) dto.Rating {
	return dto.Rating{
		Position:         position,
		UserID:           rating.UserID,
		TelegramUserID:   rating.User.TelegramUserID,
		TelegramUsername: rating.User.TelegramUsername,
		FirstName:        rating.User.FirstName,
		LastName:         rating.User.LastName,
		TotalPoints:      rating.TotalPoints,
		SolvedTasksCount: rating.SolvedTasksCount,
		LastSolvedAt:     rating.LastSolvedAt,
	}
}
