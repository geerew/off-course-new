package cron

import (
	"log/slog"

	"github.com/geerew/off-course/dao"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/utils/appFs"
	"github.com/geerew/off-course/utils/types"
	"github.com/robfig/cron/v3"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// loggerType is the type of logger
var loggerType = slog.Any("type", types.LogTypeCron)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type CronConfig struct {
	Db     database.Database
	AppFs  *appFs.AppFs
	Logger *slog.Logger
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// InitCron initializes the cron jobs
func InitCron(config *CronConfig) {
	c := cron.New()

	// Course availability
	ca := &courseAvailability{
		db:        config.Db,
		dao:       dao.NewDAO(config.Db),
		appFs:     config.AppFs,
		logger:    config.Logger,
		batchSize: 200,
	}

	go func() { ca.run() }()
	c.AddFunc("@every 5m", func() { ca.run() })

	c.Start()
}
