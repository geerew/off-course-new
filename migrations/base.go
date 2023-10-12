package migrations

import (
	"context"

	"github.com/rs/zerolog/log"

	"github.com/geerew/off-course/database"
	"github.com/uptrace/bun/migrate"
)

var Migrations = migrate.NewMigrations()

func Up(db database.Database) error {
	// Create a new migrations and discover migration files
	if err := Migrations.DiscoverCaller(); err != nil {
		return err
	}

	// Create a new migrator and init the migrations table (if not done)
	migrate := migrate.NewMigrator(db.DB(), Migrations)
	migrate.Init(context.Background())

	// Start the migration
	if err := migrate.Lock(context.Background()); err != nil {
		return err
	}
	defer migrate.Unlock(context.Background())

	group, err := migrate.Migrate(context.Background())
	if err != nil {
		return err
	}

	if group.IsZero() {
		log.Debug().Msgf("no new migrations")
		return nil
	}

	log.Info().Msgf("migrated to %s\n", group)

	return nil
}
