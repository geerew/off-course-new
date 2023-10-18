package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/geerew/off-course/api"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/migrations"
	"github.com/geerew/off-course/utils/appFs"
	"github.com/geerew/off-course/utils/jobs"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
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

	// Global logger
	cw := zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "2006-01-02T15:04:05"}
	if !*isDebug {
		log.Logger = log.Output(cw).Level(zerolog.InfoLevel)
	} else {
		log.Logger = log.Output(cw)
	}

	// Create app filesystem
	appFs := appFs.NewAppFs(afero.NewOsFs())

	// Create DB
	db := database.NewSqliteDB(&database.SqliteDbConfig{
		IsDebug: *isDebug,
		DataDir: "./co_data",
		AppFs:   appFs,
	})

	// Bootstrap the store
	if err := db.Bootstrap(); err != nil {
		log.Fatal().Err(err).Msg("")
	}

	// Run migrations
	if err := migrations.Up(db); err != nil {
		log.Fatal().Err(err).Msg("")
	}

	// Course scanner
	courseScanner := jobs.NewCourseScanner(&jobs.CourseScannerConfig{
		Db:    db,
		AppFs: appFs,
		Ctx:   context.Background(),
	})

	// Start the worker (pass in the func that will process the job)
	go courseScanner.Worker(jobs.CourseProcessor)

	// Create router
	router := api.New(&api.RouterConfig{
		Db:            db,
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
			log.Error().Err(err).Msg("")
		}
	}()

	wg.Wait()

	// TMP -> Delete all scans
	db.DB().NewDelete().Table("scans").Where("1 = 1").Exec(context.Background())
}
