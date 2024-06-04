package cron

import (
	"log/slog"
	"os"

	"github.com/geerew/off-course/daos"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/appFs"
	"github.com/geerew/off-course/utils/pagination"
)

type courseAvailability struct {
	db        database.Database
	appFs     *appFs.AppFs
	logger    *slog.Logger
	batchSize int
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (ca *courseAvailability) run() error {
	ca.logger.Debug("Updating course availability", slog.String("type", "cron"))
	perPage := 100
	page := 1
	totalPages := 1

	coursesBatch := make([]*models.Course, 0, ca.batchSize)

	courseDao := daos.NewCourseDao(ca.db)

	for page <= totalPages {
		p := pagination.New(page, perPage)
		paginationParams := &database.DatabaseParams{Pagination: p}

		// Fetch a batch of courses
		courses, err := courseDao.List(paginationParams, nil)
		if err != nil {
			attrs := []any{
				loggerType,
				slog.String("error", err.Error()),
			}

			ca.logger.Error("Failed to fetch courses", attrs...)

			return err
		}

		// Update total pages after the first fetch
		if page == 1 {
			totalPages = p.TotalPages()
		}

		// Process each course in the batch
		for _, course := range courses {
			if _, err := ca.appFs.Fs.Stat(course.Path); err != nil {
				if os.IsNotExist(err) {
					if course.Available {
						// The course is currently marked as available but is now unavailable
						course.Available = false
						coursesBatch = append(coursesBatch, course)
					}
				} else {
					// Failed to check the availability of the course
					attrs := []any{
						loggerType,
						slog.String("course", course.Title),
						slog.String("path", course.Path),
						slog.String("error", err.Error()),
					}

					ca.logger.Error("Failed to stat course", attrs...)

					return err
				}
			} else if !course.Available {
				// The course is currently marked as unavailable but is now available
				course.Available = true
				coursesBatch = append(coursesBatch, course)
			}

			// Update the courses if we hit the batch size
			if len(coursesBatch) == ca.batchSize {
				ca.writeAll(coursesBatch)
				coursesBatch = coursesBatch[:0]
			}
		}

		page++
	}

	// Update any remaining courses
	if len(coursesBatch) > 0 {
		ca.writeAll(coursesBatch)
	}

	return nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (ca *courseAvailability) writeAll(courses []*models.Course) {
	courseDao := daos.NewCourseDao(ca.db)

	// Update the courses in a transaction
	err := ca.db.RunInTransaction(func(tx *database.Tx) error {
		for _, course := range courses {
			err := courseDao.Update(course, tx)
			if err != nil {
				return err
			}

		}

		return nil
	})

	if err != nil {
		attrs := []any{
			loggerType,
			slog.String("error", err.Error()),
		}

		ca.logger.Error("Failed to update course availability", attrs...)
	}
}
