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
				DatabaseDSN:          "host=localhost dbname=redbutton",
				TasksDirectory:       "testdata/tasks",
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
