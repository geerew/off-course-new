package cron

import (
	"log/slog"
	"os"

	"github.com/geerew/off-course/daos"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/utils/pagination"
	"github.com/robfig/cron/v3"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// InitCron initializes the cron jobs
func InitCron(db database.Database, logger *slog.Logger) {
	c := cron.New()

	// Course availability check
	go func() { courseAvailability(db, logger) }()
	c.AddFunc("@every 5m", func() { courseAvailability(db, logger) })

	c.Start()
}

func courseAvailability(db database.Database, logger *slog.Logger) error {
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
			err = courseDao.Update(course, nil)
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
