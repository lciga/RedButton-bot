package logger

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"testing"
	"time"
)

func TestHandlerFormat(t *testing.T) {
	var output bytes.Buffer
	logger := New(&output, "debug")
	record := slog.NewRecord(
		time.Date(2026, time.August, 25, 23, 39, 46, 123000000, time.UTC),
		slog.LevelError,
		"Request failed",
		0,
	)
	record.Add("method", "getMe", "duration", 250*time.Millisecond)

	if err := logger.Handler().Handle(context.Background(), record); err != nil {
		t.Fatal(err)
	}

	want := "[APP] 2026/08/25 - 23:39:46.123 | ERROR | Request failed | duration=250ms | method=getMe\n"
	if output.String() != want {
		t.Fatalf("log output = %q, want %q", output.String(), want)
	}
}

func TestLoggedError(t *testing.T) {
	source := errors.New("request failed")
	marked := MarkLogged(source)
	wrapper := fmt.Errorf("initialize client: %w", marked)

	if !IsLogged(wrapper) {
		t.Fatal("wrapped marked error must be recognized as logged")
	}
	if !errors.Is(wrapper, source) {
		t.Fatal("marked error must preserve the original error chain")
	}
}
