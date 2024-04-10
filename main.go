package main

import (
	"flag"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/geerew/off-course/api"
	"github.com/geerew/off-course/daos"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/utils/appFs"
	"github.com/geerew/off-course/utils/jobs"
	"github.com/geerew/off-course/utils/pagination"
	"github.com/robfig/cron/v3"
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
		DataDir: "./oc_data",
		AppFs:   appFs,
	})

	// Bootstrap the store
	if err := db.Bootstrap(); err != nil {
		log.Fatal().Err(err).Msg("")
	}

	// Course scanner
	courseScanner := jobs.NewCourseScanner(&jobs.CourseScannerConfig{
		Db:    db,
		AppFs: appFs,
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

	// TODO: Handle this better...
	c := cron.New()
	go func() { updateCourseAvailability(db) }()
	c.AddFunc("@every 5m", func() { updateCourseAvailability(db) })
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
			log.Error().Err(err).Msg("")
		}
	}()

	wg.Wait()

	// TMP -> Delete all scans
	_, err := db.Exec("DELETE FROM " + daos.NewScanDao(db).Table)
	if err != nil {
		log.Error().Err(err).Msg("")
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func updateCourseAvailability(db database.Database) error {
	log.Info().Msg("Updating course availability...")
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
				log.Error().Err(err).Msgf("Failed to update availability for course %s", course.ID)
			}
		}

		page++
	}

	return nil
}
