/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
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
			fmt.Printf("ERR - Failed to create database manager: %s", err)
			os.Exit(1)
		}

		dao := dao.NewDAO(dbManager.DataDb)

		username, _ := cmd.Flags().GetString("user")
		if username == "" {
			fmt.Println("ERR - Username cannot be empty")
			os.Exit(1)
		}

		options := &database.Options{
			Where: squirrel.Eq{models.USER_TABLE + "." + models.USER_USERNAME: username},
		}

		user := &models.User{}
		err = dao.Get(ctx, user, options)
		if err != nil {
			if err == sql.ErrNoRows {
				fmt.Println("User not found")
				return
			}

			fmt.Printf("ERR - Failed to get user: %s", err)
			os.Exit(1)
		}

		err = dao.Delete(ctx, user, nil)
		if err != nil {
			fmt.Printf("ERR - Failed to delete user: %s", err)
			os.Exit(1)
		}

		fmt.Printf("User '%s' deleted\n", username)
	},
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func init() {
	userCmd.AddCommand(deleteCmd)
	deleteCmd.Flags().StringP("user", "u", "", "Username")
	deleteCmd.MarkFlagRequired("username")
}
