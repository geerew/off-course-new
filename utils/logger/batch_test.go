package logger

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/geerew/off-course/utils"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TesBatch_NewBatchHandler(t *testing.T) {
	t.Run("panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Expected to panic.")
			}
		}()

		NewBatchHandler(BatchOptions{})
	})

	t.Run("defaults", func(t *testing.T) {
		h := NewBatchHandler(BatchOptions{
			WriteFn: func(ctx context.Context, logs []*Log) error {
				return nil
			},
		})

		require.Equal(t, h.options.BatchSize, 100)
		require.Equal(t, h.options.Level, slog.LevelInfo)
		require.Nil(t, h.options.BeforeAddFn)
		require.NotNil(t, h.options.WriteFn)
		require.Empty(t, h.group)
		require.Empty(t, h.attrs)
		require.Empty(t, h.logs)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestBatch_LevelEnabled(t *testing.T) {
	h := NewBatchHandler(BatchOptions{
		Level: slog.LevelWarn,
		WriteFn: func(ctx context.Context, logs []*Log) error {
			return nil
		},
	})

	l := slog.New(h)

	scenarios := []struct {
		level    slog.Level
		expected bool
	}{
		{slog.LevelDebug, false},
		{slog.LevelInfo, false},
		{slog.LevelWarn, true},
		{slog.LevelError, true},
	}

	for _, s := range scenarios {
		t.Run(fmt.Sprintf("Level %v", s.level), func(t *testing.T) {
			result := l.Enabled(context.Background(), s.level)
			require.Equal(t, s.expected, result)
		})
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestBatch_SetLevel(t *testing.T) {
	h := NewBatchHandler(BatchOptions{
		Level: slog.LevelWarn,
		WriteFn: func(ctx context.Context, logs []*Log) error {
			return nil
		},
	})

	require.Equal(t, h.options.Level, slog.LevelWarn)

	h.SetLevel(slog.LevelDebug)
	require.Equal(t, h.options.Level, slog.LevelDebug)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestBatch_WithAttrsAndWithGroup(t *testing.T) {
	h0 := NewBatchHandler(BatchOptions{
		WriteFn: func(ctx context.Context, logs []*Log) error {
			return nil
		},
	})

	h1 := h0.WithAttrs([]slog.Attr{slog.Int("test1", 1)}).(*BatchHandler)
	h2 := h1.WithGroup("h2_group").(*BatchHandler)
	h3 := h2.WithAttrs([]slog.Attr{slog.Int("test2", 2)}).(*BatchHandler)

	scenarios := []struct {
		name           string
		handler        *BatchHandler
		expectedParent *BatchHandler
		expectedGroup  string
		expectedAttrs  int
	}{
		{"h0", h0, nil, "", 0},
		{"h1", h1, h0, "", 1},
		{"h2", h2, h1, "h2_group", 0},
		{"h3", h3, h2, "", 1},
	}

	for _, s := range scenarios {
		t.Run(s.name, func(t *testing.T) {
			require.Equal(t, s.expectedGroup, s.handler.group)
			require.Equal(t, s.expectedParent, s.handler.parent)
			require.Equal(t, s.expectedAttrs, len(s.handler.attrs))
		})
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestBatch_Handle(t *testing.T) {
	ctx := context.Background()

	beforeLogs := []*Log{}
	writeLogs := []*Log{}

	h := NewBatchHandler(BatchOptions{
		BatchSize: 3,
		BeforeAddFn: func(_ context.Context, log *Log) bool {
			beforeLogs = append(beforeLogs, log)
			return log.Message != "test2"
		},
		WriteFn: func(_ context.Context, logs []*Log) error {
			writeLogs = logs
			return nil
		},
	})

	h.Handle(ctx, slog.NewRecord(time.Now(), slog.LevelInfo, "test1", 0))
	h.Handle(ctx, slog.NewRecord(time.Now(), slog.LevelInfo, "test2", 0))
	h.Handle(ctx, slog.NewRecord(time.Now(), slog.LevelInfo, "test3", 0))

	// Before logs should have all the logs
	require.Equal(t, []string{"test1", "test2", "test3"}, utils.Map(beforeLogs, getMessage))

	// h.logs should have only the logs that passed the BeforeAddFn
	require.Equal(t, []string{"test1", "test3"}, utils.Map(h.logs, getMessage))

	// writeLogs should be empty because the batch size hasn't been reached
	require.Len(t, writeLogs, 0)

	// Trigger the batch write
	h.Handle(ctx, slog.NewRecord(time.Now(), slog.LevelInfo, "test4", 0))

	// h.logs should be empty because they were written
	require.Empty(t, utils.Map(h.logs, getMessage))

	// writeLogs should have the logs that passed the BeforeAddFn
	require.Equal(t, []string{"test1", "test3", "test4"}, utils.Map(writeLogs, getMessage))
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestBatch_rWriteAll(t *testing.T) {
	ctx := context.Background()

	beforeLogs := []*Log{}
	writeLogs := []*Log{}

	h := NewBatchHandler(BatchOptions{
		BatchSize: 3,
		BeforeAddFn: func(_ context.Context, log *Log) bool {
			beforeLogs = append(beforeLogs, log)
			return true
		},
		WriteFn: func(_ context.Context, logs []*Log) error {
			writeLogs = logs
			return nil
		},
	})

	h.Handle(ctx, slog.NewRecord(time.Now(), slog.LevelInfo, "test1", 0))
	h.Handle(ctx, slog.NewRecord(time.Now(), slog.LevelInfo, "test2", 0))

	require.Equal(t, []string{"test1", "test2"}, utils.Map(beforeLogs, getMessage))
	require.Equal(t, []string{"test1", "test2"}, utils.Map(h.logs, getMessage))
	require.Empty(t, utils.Map(writeLogs, getMessage))

	h.WriteAll(ctx)

	require.Equal(t, []string{"test1", "test2"}, utils.Map(beforeLogs, getMessage))
	require.Empty(t, utils.Map(h.logs, getMessage))
	require.Equal(t, []string{"test1", "test2"}, utils.Map(writeLogs, getMessage))
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestBatch_AttrsFormat(t *testing.T) {
	ctx := context.Background()

	beforeLogs := []*Log{}

	h0 := NewBatchHandler(BatchOptions{
		BeforeAddFn: func(_ context.Context, log *Log) bool {
			beforeLogs = append(beforeLogs, log)
			return true
		},
		WriteFn: func(_ context.Context, logs []*Log) error {
			return nil
		},
	})

	h1 := h0.WithAttrs([]slog.Attr{slog.Int("a", 1), slog.String("b", "123")})

	h2 := h1.WithGroup("sub").WithAttrs([]slog.Attr{
		slog.Int("c", 3),
		slog.Any("d", map[string]any{"d.1": 1}),
		slog.Any("e", errors.New("example error")),
	})

	record := slog.NewRecord(time.Now(), slog.LevelInfo, "hello", 0)
	record.AddAttrs(slog.String("name", "test"))

	h0.Handle(ctx, record)
	h1.Handle(ctx, record)
	h2.Handle(ctx, record)

	expected := []string{
		`{"name":"test"}`,
		`{"a":1,"b":"123","name":"test"}`,
		`{"a":1,"b":"123","sub":{"c":3,"d":{"d.1":1},"e":"example error","name":"test"}}`,
	}

	for i, ex := range expected {
		t.Run(fmt.Sprintf("log handler %d", i), func(t *testing.T) {
			log := beforeLogs[i]
			raw, _ := log.Data.MarshalJSON()
			require.JSONEq(t, ex, string(raw))
		})
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// getMessage is a helper function that returns the message of a Log
func getMessage(l *Log) string {
	return l.Message
}
