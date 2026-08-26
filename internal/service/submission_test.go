package service

import (
	"testing"

	"RedButton-bot/internal/database/model"
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
