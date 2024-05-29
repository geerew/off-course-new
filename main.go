package main

import (
	"context"
	"database/sql"
	"flag"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/geerew/off-course/api"
	"github.com/geerew/off-course/daos"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/appFs"
	"github.com/geerew/off-course/utils/jobs"
	"github.com/geerew/off-course/utils/logger"
	"github.com/geerew/off-course/utils/pagination"
	"github.com/geerew/off-course/utils/security"
	"github.com/geerew/off-course/utils/types"
	"github.com/robfig/cron/v3"
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
	appFs := appFs.NewAppFs(afero.NewOsFs())

	// Create the database manager, which will create the data/logs databases
	dbManager, err := database.NewDBManager(&database.DatabaseConfig{
		IsDebug:  *isDebug,
		DataDir:  "./oc_data",
		AppFs:    appFs,
		InMemory: false,
	})

	if err != nil {
		log.Fatal("Failed to create database manager", err)
	}

	// Logger
	logger, err := logger.InitLogger(loggerWriteFn(dbManager.LogsDb))
	if err != nil {
		log.Fatal("Failed to initialize logger", err)
	}

	// Course scanner
	courseScanner := jobs.NewCourseScanner(&jobs.CourseScannerConfig{
		Db:    dbManager.DataDb,
		AppFs: appFs,
	})

	// Start the worker (pass in the func that will process the job)
	go courseScanner.Worker(jobs.CourseProcessor)

	// Create router
	router := api.New(&api.RouterConfig{
		DbManager:     dbManager,
		Logger:        logger,
		AppFs:         appFs,
		CourseScanner: courseScanner,
		Port:          *port,
		IsProduction:  isProduction,
	})

	// TODO: Handle this better...
	c := cron.New()
	go func() { updateCourseAvailability(dbManager.DataDb, logger) }()
	c.AddFunc("@every 5m", func() { updateCourseAvailability(dbManager.DataDb, logger) })
	c.Start()

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

	// TMP -> Delete all scans
	_, err = dbManager.DataDb.Exec("DELETE FROM " + daos.NewScanDao(dbManager.DataDb).Table())
	if err != nil {
		log.Fatal("Failed to delete scans", err)
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func updateCourseAvailability(db database.Database, logger *slog.Logger) error {
	logger.Debug("Updating course availability", slog.String("type", "cron"))
	const perPage = 100
	var page = 1

	// This will be updated after the first fetch
	var totalPages = 1

	courseDao := daos.NewCourseDao(db)

	for page <= totalPages {
		p := pagination.New(page, perPage)
		paginationParams := &database.DatabaseParams{Pagination: p}

		// Fetch a batch of courses
		courses, err := courseDao.List(paginationParams, nil)
		if err != nil {
			return err
		}

		// Update total pages after the first fetch
		if page == 1 {
			totalPages = p.TotalPages()
		}

		// Process each course in the batch
		for _, course := range courses {
			course.Available = false
			_, err := os.Stat(course.Path)
			if err == nil {
				course.Available = true
			}

			// Update the course's availability in the database
			err = courseDao.Update(course)
			if err != nil {
				attrs := []any{
					slog.String("type", "cron"),
					slog.String("course", course.Title),
					slog.String("error", err.Error()),
				}

				logger.Error("Failed to update availability for course", attrs...)
			}
		}

		page++
	}

	return nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// loggerWriteFn returns a logger.WriteFn that writes logs to the database
func loggerWriteFn(db database.Database) logger.WriteFn {
	return func(ctx context.Context, logs []*logger.Log) error {

		// Write accumulated logs
		db.RunInTransaction(func(tx *sql.Tx) error {
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
