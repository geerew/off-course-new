package logger

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/geerew/off-course/utils/types"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type Log struct {
	Time    time.Time
	Message string
	Level   slog.Level
	Data    types.JsonMap
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func InitLogger(writeFn WriteFn) (*slog.Logger, error) {
	duration := 3 * time.Second
	ticker := time.NewTicker(duration)
	done := make(chan bool)

	handler := NewBatchHandler(BatchOptions{
		Level:     slog.LevelDebug,
		BatchSize: 200,
		WriteFn:   writeFn,
	})

	go func() {
		ctx := context.Background()

		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				handler.WriteAll(ctx)
			}
		}
	}()

	logger := slog.New(handler)

	return logger, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func BasicWriteFn() WriteFn {
	return func(ctx context.Context, logs []*Log) error {
		for _, l := range logs {
			fmt.Println(l.Level, l.Message, l.Data)
		}
		return nil
	}
}
