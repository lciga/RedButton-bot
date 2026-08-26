package config

import (
	"os"
	"reflect"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name       string
		env        map[string]string
		createFile bool
		want       *Config
		wantErr    bool
	}{
		{
			name: "valid configuration",
			env: map[string]string{
				"TELEGRAM_BOT_TOKEN":    "token",
				"DATABASE_DSN":          "host=localhost dbname=redbutton",
				"TASKS_DIRECTORY":       "testdata/tasks",
				"TASK_START_DATE":       "2026-09-01T00:00:00",
				"TASK_END_DATE":         "2026-09-07T00:00:00",
				"TASK_EXPIRE":           "24h",
				"NOTIFICATION_INTERVAL": "10s",
				"ADMIN_TELEGRAM_IDS":    "123456789, 987654321",
			},
			want: &Config{
				TelegramBotToken:     "token",
				TelegramInitTimeout:  30 * time.Second,
				DatabaseDSN:          "host=localhost dbname=redbutton",
				TasksDirectory:       "testdata/tasks",
				TimeZone:             utcPlusFive,
				BotStartDate:         time.Date(2026, time.September, 1, 0, 0, 0, 0, utcPlusFive),
				BotEndDate:           time.Date(2026, time.September, 7, 0, 0, 0, 0, utcPlusFive),
				TaskExpire:           24 * time.Hour,
				NotificationInterval: 10 * time.Second,
				AdminTelegramIDs: map[int64]struct{}{
					123456789: {},
					987654321: {},
				},
			},
		},
		{
			name: "custom timezone",
			env: map[string]string{
				"TASK_TIMEZONE":   "Europe/Moscow",
				"TASK_START_DATE": "2026-09-01T00:00:00",
				"TASK_END_DATE":   "2026-09-07T00:00:00",
				"TASK_EXPIRE":     "24h",
			},
			createFile: true,
			want: &Config{
				TelegramInitTimeout:  30 * time.Second,
				TasksDirectory:       "tasks",
				TimeZone:             mustLoadLocation(t, "Europe/Moscow"),
				BotStartDate:         time.Date(2026, time.September, 1, 0, 0, 0, 0, mustLoadLocation(t, "Europe/Moscow")),
				BotEndDate:           time.Date(2026, time.September, 7, 0, 0, 0, 0, mustLoadLocation(t, "Europe/Moscow")),
				TaskExpire:           24 * time.Hour,
				NotificationInterval: 15 * time.Second,
				AdminTelegramIDs:     map[int64]struct{}{},
			},
		},
		{
			name: "invalid timezone",
			env: map[string]string{
				"TASK_TIMEZONE": "not/a-timezone",
			},
			createFile: true,
			wantErr:    true,
		},
		{
			name: "invalid admin id",
			env: map[string]string{
				"TASK_START_DATE":    "2026-09-01T00:00:00",
				"TASK_END_DATE":      "2026-09-07T00:00:00",
				"TASK_EXPIRE":        "24h",
				"ADMIN_TELEGRAM_IDS": "invalid",
			},
			createFile: true,
			wantErr:    true,
		},
		{
			name: "invalid bot period",
			env: map[string]string{
				"TASK_START_DATE": "2026-09-07T00:00:00",
				"TASK_END_DATE":   "2026-09-01T00:00:00",
				"TASK_EXPIRE":     "24h",
			},
			createFile: true,
			wantErr:    true,
		},
		{
			name:    "missing dotenv file",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalDir, err := os.Getwd()
			if err != nil {
				t.Fatal(err)
			}
			t.Chdir(t.TempDir())
			t.Cleanup(func() {
				if err := os.Chdir(originalDir); err != nil {
					t.Errorf("restore working directory: %v", err)
				}
			})
			if tt.createFile {
				if err := os.WriteFile(".env", nil, 0o600); err != nil {
					t.Fatal(err)
				}
			}
			for key, value := range tt.env {
				t.Setenv(key, value)
			}

			got, err := Load()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Load() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Load() = %v, want %v", got, tt.want)
			}
		})
	}
}

func mustLoadLocation(t *testing.T, name string) *time.Location {
	t.Helper()
	location, err := time.LoadLocation(name)
	if err != nil {
		t.Fatal(err)
	}
	return location
}

func TestLoadTimeZoneUTCOffset(t *testing.T) {
	location, err := loadTimeZone("UTC+05:30")
	if err != nil {
		t.Fatal(err)
	}
	_, offset := time.Date(2026, time.January, 1, 0, 0, 0, 0, location).Zone()
	if offset != 5*60*60+30*60 {
		t.Errorf("offset = %d, want %d", offset, 5*60*60+30*60)
	}
}

func TestLoadTimeZoneRejectsInvalidOffsets(t *testing.T) {
	for _, value := range []string{"UTC+15", "UTC-14:01", "UTC+05:60"} {
		t.Run(value, func(t *testing.T) {
			if _, err := loadTimeZone(value); err == nil {
				t.Fatalf("loadTimeZone(%q) returned no error", value)
			}
		})
	}
}
