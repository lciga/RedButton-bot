package bot

import (
	"testing"
	"time"
)

func TestNotificationDelay(t *testing.T) {
	now := time.Date(2026, time.September, 1, 11, 59, 0, 0, time.UTC)
	activation := now.Add(time.Minute)

	if got := notificationDelay(now, activation); got != time.Minute {
		t.Errorf("notificationDelay() = %v, want %v", got, time.Minute)
	}
	if got := notificationDelay(now, now.Add(-time.Minute)); got != 0 {
		t.Errorf("notificationDelay() = %v, want 0", got)
	}
}
