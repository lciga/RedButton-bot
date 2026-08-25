package logger

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

type gormHandler struct {
	logger        *slog.Logger
	level         gormlogger.LogLevel
	slowThreshold time.Duration
}

// Функция создания адаптера логгера для GORM
func GORM(logger *slog.Logger) gormlogger.Interface {
	return &gormHandler{
		logger:        logger.With("component", "database"),
		level:         gormlogger.Warn,
		slowThreshold: 500 * time.Millisecond,
	}
}

func (h *gormHandler) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	clone := *h
	clone.level = level
	return &clone
}

func (h *gormHandler) Info(ctx context.Context, message string, values ...any) {
	if h.level >= gormlogger.Info {
		h.logger.DebugContext(ctx, fmt.Sprintf(message, values...))
	}
}

func (h *gormHandler) Warn(ctx context.Context, message string, values ...any) {
	if h.level >= gormlogger.Warn {
		h.logger.WarnContext(ctx, fmt.Sprintf(message, values...))
	}
}

func (h *gormHandler) Error(ctx context.Context, message string, values ...any) {
	if h.level >= gormlogger.Error {
		h.logger.ErrorContext(ctx, fmt.Sprintf(message, values...))
	}
}

func (h *gormHandler) Trace(
	ctx context.Context,
	startedAt time.Time,
	query func() (string, int64),
	err error,
) {
	if h.level == gormlogger.Silent || errors.Is(err, gorm.ErrRecordNotFound) {
		return
	}

	elapsed := time.Since(startedAt)
	sql, rows := query()
	attributes := []any{
		"elapsed", elapsed,
		"rows", rows,
		"query", sql,
	}
	switch {
	case err != nil && h.level >= gormlogger.Error:
		h.logger.ErrorContext(ctx, "Database query failed", append(attributes, "error", err)...)
	case elapsed >= h.slowThreshold && h.level >= gormlogger.Warn:
		h.logger.WarnContext(ctx, "Slow database query", attributes...)
	case h.level >= gormlogger.Info:
		h.logger.DebugContext(ctx, "Database query completed", attributes...)
	}
}
