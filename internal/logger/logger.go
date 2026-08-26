// Пакет для работы с логами
package logger

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"sort"
	"strings"
	"sync"
	"time"
)

type loggedError struct {
	err error
}

func (e *loggedError) Error() string {
	return e.err.Error()
}

func (e *loggedError) Unwrap() error {
	return e.err
}

// Функция отметки уже записанной в лог ошибки
func MarkLogged(err error) error {
	if err == nil || IsLogged(err) {
		return err
	}
	return &loggedError{err: err}
}

// Функция проверки наличия ошибки в логах
func IsLogged(err error) bool {
	var target *loggedError
	return errors.As(err, &target)
}

// Структура форматтера логов
type handler struct {
	writer io.Writer
	level  slog.Level
	attrs  []slog.Attr
	groups []string
	mutex  *sync.Mutex
}

// Функция создания логгера
func New(writer io.Writer, levelValue string) *slog.Logger {
	level := slog.LevelInfo
	if err := level.UnmarshalText([]byte(strings.ToLower(levelValue))); err != nil {
		level = slog.LevelInfo
	}

	return slog.New(&handler{
		writer: writer,
		level:  level,
		mutex:  &sync.Mutex{},
	})
}

func (h *handler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *handler) Handle(_ context.Context, record slog.Record) error {
	attrs := make([]slog.Attr, 0, len(h.attrs)+record.NumAttrs())
	attrs = append(attrs, h.attrs...)
	record.Attrs(func(attr slog.Attr) bool {
		attrs = append(attrs, attr)
		return true
	})
	sort.SliceStable(attrs, func(left, right int) bool {
		return attrs[left].Key < attrs[right].Key
	})

	var builder strings.Builder
	fmt.Fprintf(
		&builder,
		"[APP] %s | %s | %s",
		record.Time.Format("2006/01/02 - 15:04:05.000"),
		strings.ToUpper(record.Level.String()),
		record.Message,
	)
	for _, attr := range attrs {
		writeAttr(&builder, h.groups, attr)
	}
	builder.WriteByte('\n')

	h.mutex.Lock()
	defer h.mutex.Unlock()
	_, err := io.WriteString(h.writer, builder.String())
	return err
}

func (h *handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	clone := *h
	clone.attrs = append(append([]slog.Attr(nil), h.attrs...), attrs...)
	return &clone
}

func (h *handler) WithGroup(name string) slog.Handler {
	clone := *h
	clone.groups = append(append([]string(nil), h.groups...), name)
	return &clone
}

func writeAttr(builder *strings.Builder, groups []string, attr slog.Attr) {
	attr.Value = attr.Value.Resolve()
	if attr.Equal(slog.Attr{}) {
		return
	}
	if attr.Value.Kind() == slog.KindGroup {
		groups = append(groups, attr.Key)
		for _, child := range attr.Value.Group() {
			writeAttr(builder, groups, child)
		}
		return
	}

	key := strings.Join(append(groups, attr.Key), ".")
	fmt.Fprintf(builder, " | %s=%s", key, attrValue(attr.Value))
}

func attrValue(value slog.Value) string {
	switch value.Kind() {
	case slog.KindTime:
		return value.Time().Format(time.RFC3339Nano)
	case slog.KindDuration:
		return value.Duration().String()
	default:
		return fmt.Sprint(value.Any())
	}
}
