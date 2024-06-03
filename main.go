package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/geerew/off-course/api"
	"github.com/geerew/off-course/cron"
	"github.com/geerew/off-course/daos"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/appFs"
	"github.com/geerew/off-course/utils/jobs"
	"github.com/geerew/off-course/utils/logger"
	"github.com/geerew/off-course/utils/security"
	"github.com/geerew/off-course/utils/types"
	"github.com/spf13/afero"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

var isProduction = !strings.HasPrefix(os.Args[0], os.TempDir())

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func main() {
	// Flags
	port := flag.String("port", ":9081", "server port")
	isDebug := flag.Bool("debug", false, "verbose")
	flag.Parse()

	// Create app filesystem
	appFs := appFs.NewAppFs(afero.NewOsFs(), nil)

	// Create the database manager, which will create the data/logs databases
	dbManager, err := database.NewSqliteDBManager(&database.DatabaseConfig{
		IsDebug:  *isDebug,
		DataDir:  "./oc_data",
		AppFs:    appFs,
		InMemory: false,
	})
	if err != nil {
		log.Fatal("Failed to create database manager", err)
	}

	// Logger
	logger, loggerDone, err := logger.InitLogger(&logger.BatchOptions{
		BatchSize:   200,
		BeforeAddFn: loggerBeforeAddFunc(dbManager.LogsDb),
		WriteFn:     loggerWriteFn(dbManager.LogsDb),
	})

	if err != nil {
		log.Fatal("Failed to initialize logger", err)
	}

	// Set loggers
	dbManager.DataDb.SetLogger(logger)
	appFs.SetLogger(logger)

	// Course scanner
	courseScanner := jobs.NewCourseScanner(&jobs.CourseScannerConfig{
		Db:     dbManager.DataDb,
		AppFs:  appFs,
		Logger: logger,
	})

	// Start the worker (pass in the func that will process the job)
	go courseScanner.Worker(jobs.CourseProcessor, nil)

	// Initialize cron jobs
	cron.InitCron(dbManager.DataDb, logger)

	// Create router
	router := api.New(&api.RouterConfig{
		DbManager:     dbManager,
		Logger:        logger,
		AppFs:         appFs,
		CourseScanner: courseScanner,
		Port:          *port,
		IsProduction:  isProduction,
	})

	var wg sync.WaitGroup
	wg.Add(1)

	// Wait for interrupt signal, to gracefully shutdown
	go func() {
		defer wg.Done()
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
		<-quit
	}()

	// Serve the UI/API
	go func() {
		defer wg.Done()
		if err := router.Serve(); err != nil {
			log.Fatal("Failed to start router", err)
		}
	}()

	wg.Wait()

	// Delete all scans
	_, err = dbManager.DataDb.Exec("DELETE FROM " + daos.NewScanDao(dbManager.DataDb).Table())
	if err != nil {
		log.Fatal("Failed to delete scans", err)
	}

	// Close the logger, which will write any remaining logs
	close(loggerDone)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// loggerBeforeAddFunc is a logger.BeforeAddFn
func loggerBeforeAddFunc(db database.Database) logger.BeforeAddFn {
	logsDao := daos.NewLogDao(db)

	return func(ctx context.Context, log *logger.Log) bool {
		// Skip calls to the logs API
		if strings.HasPrefix(log.Message, "GET /api/logs/") {
			return false
		}

		// This should never happen as the logsDb should be nil, but in the event it is not, skip
		// logging log writes as it will cause an infinite loop
		if strings.HasPrefix(log.Message, "INSERT INTO "+logsDao.Table()) ||
			strings.HasPrefix(log.Message, "SELECT "+logsDao.Table()) ||
			strings.HasPrefix(log.Message, "UPDATE "+logsDao.Table()) ||
			strings.HasPrefix(log.Message, "DELETE FROM "+logsDao.Table()) {
			return false
		}

		return true
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// loggerWriteFn returns a logger.WriteFn that writes logs to the database
func loggerWriteFn(db database.Database) logger.WriteFn {
	return func(ctx context.Context, logs []*logger.Log) error {
		// Write accumulated logs
		db.RunInTransaction(func(tx *database.Tx) error {
			model := &models.Log{}

			for _, l := range logs {
				model.ID = security.PseudorandomString(10)
				model.Level = int(l.Level)
				model.Message = l.Message
				model.Data = l.Data
				model.CreatedAt, _ = types.ParseDateTime(l.Time)
				model.UpdatedAt = model.CreatedAt

				if err := daos.NewLogDao(db).Write(model, tx); err != nil {
					log.Println("Failed to write log", model, err)
				}
			}

			return nil
		})

		return nil
	}
}
