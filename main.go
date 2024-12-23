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
	"github.com/geerew/off-course/dao"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils"
	"github.com/geerew/off-course/utils/appFs"
	"github.com/geerew/off-course/utils/auth"
	"github.com/geerew/off-course/utils/coursescan"
	"github.com/geerew/off-course/utils/logger"
	"github.com/geerew/off-course/utils/security"
	"github.com/spf13/afero"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func main() {
	// Flags
	http := flag.String("http", "127.0.0.1:9081", "TCP address to listen for the HTTP server")
	devMode := flag.Bool("dev", false, "development mode")
	flag.Parse()

	ctx := context.Background()

	appFs := appFs.NewAppFs(afero.NewOsFs(), nil)

	dbManager, err := database.NewSqliteDBManager(&database.DatabaseConfig{
		DataDir:  "./oc_data",
		AppFs:    appFs,
		InMemory: false,
	})

	if err != nil {
		log.Fatal("Failed to create database manager:", err)
	}

	logger, loggerDone, err := logger.InitLogger(&logger.BatchOptions{
		BatchSize:   200,
		BeforeAddFn: loggerBeforeAddFn(dbManager.LogsDb),
		WriteFn:     loggerWriteFn(ctx, dbManager.LogsDb),
	})

	if err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}
	defer close(loggerDone)

	// Set DB loggers
	dbManager.DataDb.SetLogger(logger)
	appFs.SetLogger(logger)

	courseScan := coursescan.NewCourseScan(&coursescan.CourseScanConfig{
		Db:     dbManager.DataDb,
		AppFs:  appFs,
		Logger: logger,
	})

	// Start the course scan worker
	go courseScan.Worker(ctx, coursescan.Processor, nil)

	cron.InitCron(&cron.CronConfig{
		Db:     dbManager.DataDb,
		AppFs:  appFs,
		Logger: logger,
	})

	jwtSecret, err := auth.GetJwtSecret(dbManager.DataDb)
	if err != nil {
		utils.Errf("Failed to get JWT secret: %s", err)
		os.Exit(1)
	}

	router := api.NewRouter(&api.RouterConfig{
		DbManager:    dbManager,
		Logger:       logger,
		AppFs:        appFs,
		CourseScan:   courseScan,
		HttpAddr:     *http,
		IsProduction: !*devMode,
		JwtSecret:    jwtSecret,
	})

	var wg sync.WaitGroup
	wg.Add(1)

	// Listen for shutdown signals
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
			utils.Errf("Failed to start router:: %s", err)
			os.Exit(1)
		}
	}()

	wg.Wait()

	utils.Infof("Shutting down...")

	// Delete all scans
	_, err = dbManager.DataDb.Exec("DELETE FROM " + models.SCAN_TABLE)
	if err != nil {
		utils.Errf("Failed to delete scans: %s", err)
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// loggerBeforeAddFunc is a logger.BeforeAddFn
func loggerBeforeAddFn(db database.Database) logger.BeforeAddFn {
	return func(ctx context.Context, log *logger.Log) bool {
		// Skip calls to the logs API
		if strings.HasPrefix(log.Message, "GET /api/logs") {
			return false
		}

		// This should never happen as the logsDb should be nil, but in the event it is not, skip
		// logging log writes as it will cause an infinite loop
		if strings.HasPrefix(log.Message, "INSERT INTO "+models.LOG_TABLE) ||
			strings.HasPrefix(log.Message, "SELECT "+models.LOG_TABLE) ||
			strings.HasPrefix(log.Message, "UPDATE "+models.LOG_TABLE) ||
			strings.HasPrefix(log.Message, "DELETE FROM "+models.LOG_TABLE) {
			return false
		}

		return true
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// loggerWriteFn returns a logger.WriteFn that writes logs to the database
func loggerWriteFn(ctx context.Context, db database.Database) logger.WriteFn {
	return func(ctx context.Context, logs []*logger.Log) error {
		logDao := dao.NewDAO(db)

		// Write accumulated logs
		db.RunInTransaction(ctx, func(txCtx context.Context) error {
			model := &models.Log{}

			for _, l := range logs {
				model.ID = security.PseudorandomString(10)
				model.Level = int(l.Level)
				model.Message = l.Message
				model.Data = l.Data
				model.CreatedAt = l.Time
				model.UpdatedAt = model.CreatedAt

				// Write the log
				err := logDao.WriteLog(txCtx, model)
				if err != nil {
					log.Println("Failed to write log", model, err)
				}
			}

			return nil
		})

		return nil
	}
}
