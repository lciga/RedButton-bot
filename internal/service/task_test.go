package service

import (
	"testing"
	"time"
)

func TestCalculateTaskEndAt(t *testing.T) {
	startsAt := time.Date(2026, time.September, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		expire    time.Duration
		botEndsAt time.Time
		want      time.Time
	}{
		{
			name:      "task expire",
			expire:    24 * time.Hour,
			botEndsAt: startsAt.Add(7 * 24 * time.Hour),
			want:      startsAt.Add(24 * time.Hour),
		},
		{
			name:      "bot end",
			expire:    24 * time.Hour,
			botEndsAt: startsAt.Add(6 * time.Hour),
			want:      startsAt.Add(6 * time.Hour),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := calculateTaskEndAt(startsAt, tt.expire, tt.botEndsAt); !got.Equal(tt.want) {
				t.Errorf("calculateTaskEndAt() = %v, want %v", got, tt.want)
			}
		})
	}
}
