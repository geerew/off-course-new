package logger

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
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

// InitLogger initializes a logger with a batch handler
//
// During testing, the batchSize should be set to 1 to ensure that logs are written immediately
func InitLogger(writeFn WriteFn, batchSize int) (*slog.Logger, chan bool, error) {
	duration := 3 * time.Second
	ticker := time.NewTicker(duration)
	done := make(chan bool)

	handler := NewBatchHandler(BatchOptions{
		Level:     slog.LevelDebug,
		BatchSize: batchSize,
		WriteFn:   writeFn,
	})

	go func() {
		ctx := context.Background()

		for {
			select {
			case <-done:
				handler.WriteAll(ctx)
				return
			case <-ticker.C:
				handler.WriteAll(ctx)
			}
		}
	}()

	logger := slog.New(handler)

	return logger, done, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// BasicWriteFn is a WriteFn that writes logs to stdout
func BasicWriteFn() WriteFn {
	return func(ctx context.Context, logs []*Log) error {
		for _, l := range logs {
			fmt.Println(l.Level, l.Message, l.Data)
		}
		return nil
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NilWriteFn is a WriteFn that does nothing
func NilWriteFn() WriteFn {
	return func(ctx context.Context, logs []*Log) error {
		return nil
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// TestWriteFn is a WriteFn that writes logs to a slice. For use in tests
//
// Example:
//
//	var logs []*logger.Log
//	var logsMux sync.Mutex
//	loggy, done, err := logger.InitLogger(logger.TestWriteFn(&logs, &logsMux), 1)
func TestWriteFn(logs *[]*Log, mux *sync.Mutex) WriteFn {
	return func(ctx context.Context, newLogs []*Log) error {
		mux.Lock()
		defer mux.Unlock()
		*logs = append(*logs, newLogs...)
		return nil
	}
}
