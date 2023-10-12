package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/migrations"
	"github.com/geerew/off-course/utils/appFs"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/afero"
	"github.com/uptrace/bun/migrate"
	"github.com/urfave/cli/v2"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

var isProduction = !strings.HasPrefix(os.Args[0], os.TempDir())

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func main() {
	// Flags
	// port := flag.String("port", ":9081", "server port")
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

	if err := migrations.Up(db); err != nil {
		log.Fatal().Err(err).Msg("")
	}

	// // DB migrations
	// // if err := models.MigrateModels(db); err != nil {
	// // log.Fatal().Err(err).Msg("")
	// // }

	// // Course scanner
	// courseScanner := jobs.NewCourseScanner(&jobs.CourseScannerConfig{
	// 	Db:    db,
	// 	AppFs: appFs,
	// })

	// // Start the worker (pass in the func that will process the job)
	// go courseScanner.Worker(jobs.CourseProcessor)

	// Create router
	// router := api.New(&api.RouterConfig{
	// 	Db:            db,
	// 	AppFs:         appFs,
	// 	CourseScanner: nil,
	// 	Port:          *port,
	// 	IsProduction:  isProduction,
	// })

	// var wg sync.WaitGroup
	// wg.Add(1)

	// // Wait for interrupt signal, to gracefully shutdown
	// go func() {
	// 	defer wg.Done()
	// 	quit := make(chan os.Signal, 1)
	// 	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	// 	<-quit
	// }()

	// // Serve the UI/API
	// go func() {
	// 	defer wg.Done()
	// 	if err := router.Serve(); err != nil {
	// 		log.Error().Err(err).Msg("")
	// 	}
	// }()

	// wg.Wait()

	// TMP -> Delete all scans
	// db.DB().Where("1 = 1").Unscoped().Delete(&models.Scan{})
}

func newDBCommand(migrator *migrate.Migrator) *cli.Command {
	return &cli.Command{
		Name:  "db",
		Usage: "database migrations",
		Subcommands: []*cli.Command{
			{
				Name:  "init",
				Usage: "create migration tables",
				Action: func(c *cli.Context) error {
					return migrator.Init(c.Context)
				},
			},
			{
				Name:  "migrate",
				Usage: "migrate database",
				Action: func(c *cli.Context) error {
					if err := migrator.Lock(c.Context); err != nil {
						return err
					}
					defer migrator.Unlock(c.Context) //nolint:errcheck

					group, err := migrator.Migrate(c.Context)
					if err != nil {
						return err
					}
					if group.IsZero() {
						fmt.Printf("there are no new migrations to run (database is up to date)\n")
						return nil
					}
					fmt.Printf("migrated to %s\n", group)
					return nil
				},
			},
			{
				Name:  "rollback",
				Usage: "rollback the last migration group",
				Action: func(c *cli.Context) error {
					if err := migrator.Lock(c.Context); err != nil {
						return err
					}
					defer migrator.Unlock(c.Context) //nolint:errcheck

					group, err := migrator.Rollback(c.Context)
					if err != nil {
						return err
					}
					if group.IsZero() {
						fmt.Printf("there are no groups to roll back\n")
						return nil
					}
					fmt.Printf("rolled back %s\n", group)
					return nil
				},
			},
			{
				Name:  "lock",
				Usage: "lock migrations",
				Action: func(c *cli.Context) error {
					return migrator.Lock(c.Context)
				},
			},
			{
				Name:  "unlock",
				Usage: "unlock migrations",
				Action: func(c *cli.Context) error {
					return migrator.Unlock(c.Context)
				},
			},
			{
				Name:  "create_go",
				Usage: "create Go migration",
				Action: func(c *cli.Context) error {
					name := strings.Join(c.Args().Slice(), "_")
					mf, err := migrator.CreateGoMigration(c.Context, name)
					if err != nil {
						return err
					}
					fmt.Printf("created migration %s (%s)\n", mf.Name, mf.Path)
					return nil
				},
			},
			{
				Name:  "create_sql",
				Usage: "create up and down SQL migrations",
				Action: func(c *cli.Context) error {
					name := strings.Join(c.Args().Slice(), "_")
					files, err := migrator.CreateSQLMigrations(c.Context, name)
					if err != nil {
						return err
					}

					for _, mf := range files {
						fmt.Printf("created migration %s (%s)\n", mf.Name, mf.Path)
					}

					return nil
				},
			},
			{
				Name:  "status",
				Usage: "print migrations status",
				Action: func(c *cli.Context) error {
					ms, err := migrator.MigrationsWithStatus(c.Context)
					if err != nil {
						return err
					}
					fmt.Printf("migrations: %s\n", ms)
					fmt.Printf("unapplied migrations: %s\n", ms.Unapplied())
					fmt.Printf("last migration group: %s\n", ms.LastGroup())
					return nil
				},
			},
			{
				Name:  "mark_applied",
				Usage: "mark migrations as applied without actually running them",
				Action: func(c *cli.Context) error {
					group, err := migrator.Migrate(c.Context, migrate.WithNopMigration())
					if err != nil {
						return err
					}
					if group.IsZero() {
						fmt.Printf("there are no new migrations to mark as applied\n")
						return nil
					}
					fmt.Printf("marked as applied %s\n", group)
					return nil
				},
			},
		},
	}
}
