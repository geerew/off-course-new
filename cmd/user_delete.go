package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/dao"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/appFs"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

// TODO: Delete all sessions when deleting a user

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a user",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println()

		ctx := context.Background()
		appFs := appFs.NewAppFs(afero.NewOsFs(), nil)

		dbManager, err := database.NewSqliteDBManager(&database.DatabaseConfig{
			DataDir:  "./oc_data",
			AppFs:    appFs,
			InMemory: false,
		})

		if err != nil {
			errorMessage("Failed to create database manager: %s", err)
			os.Exit(1)
		}

		var username string
		for {
			username = questionPlain("Username")
			if username != "" {
				break
			}

			errorMessage("Username cannot be empty")
		}

		fmt.Println()

		dao := dao.NewDAO(dbManager.DataDb)
		options := &database.Options{
			Where: squirrel.Eq{models.USER_TABLE + "." + models.USER_USERNAME: username},
		}

		user := &models.User{}

		err = dao.Get(ctx, user, options)
		if err != nil {
			if err == sql.ErrNoRows {
				errorMessage("User '%s' not found\n", username)
				return
			}

			errorMessage("Failed to lookup user: %s", err)
			os.Exit(1)
		}

		err = dao.Delete(ctx, user, nil)
		if err != nil {
			errorMessage("Failed to delete user: %s", err)
			os.Exit(1)
		}

		successMessage("User '%s' deleted", username)
	},
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func init() {
	userCmd.AddCommand(deleteCmd)

	// May add flags in the future to allow headless
	// deleteCmd.Flags().StringP("user", "u", "", "Username")
	// deleteCmd.MarkFlagRequired("username")
}
