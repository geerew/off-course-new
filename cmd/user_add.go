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
	"github.com/geerew/off-course/utils/auth"
	"github.com/geerew/off-course/utils/types"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a user",
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

		var username string
		for {
			username = questionPlain("Username")
			if username != "" {
				break
			}

			errorMessage("Username cannot be empty")
		}

		var role string
		for {
			role = questionPlain("Role (admin/user)")
			if role == "admin" || role == "user" {
				break
			}

			errorMessage("Invalid role. Must be 'admin' or 'user'")
		}

		// Get password
		var password string
		for {
			password = questionPassword("Password")
			if password != "" {
				break
			}

			errorMessage("Password cannot be empty")
		}

		// Confirm password
		for {
			pwd := questionPassword("Confirm Password")
			if pwd == password {
				break
			}

			errorMessage("Passwords do not match")
		}

		fmt.Println()

		user := &models.User{
			Username:     username,
			PasswordHash: auth.GeneratePassword(password),
		}

		if role == "admin" {
			user.Role = types.UserRoleAdmin
		} else {
			user.Role = types.UserRoleUser
		}

		dao := dao.NewDAO(dbManager.DataDb)
		options := &database.Options{
			Where: squirrel.Eq{models.USER_TABLE + "." + models.USER_USERNAME: username},
		}

		err = dao.Get(ctx, user, options)
		if err != nil && err != sql.ErrNoRows {
			errorMessage("Failed to lookup user: %s", err)
			os.Exit(1)
		}

		if err == nil {
			errorMessage("Username '%s' already exists", username)
			os.Exit(1)
		}

		err = dao.CreateUser(ctx, user)
		if err != nil {
			errorMessage("Failed to create user: %s", err)
			os.Exit(1)
		}

		successMessage("User '%s' created\n", username)
	},
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func init() {
	userCmd.AddCommand(addCmd)

	// May add flags in the future to allow headless
	// 	addCmd.Flags().StringP("user", "u", "", "Username")
	// 	addCmd.Flags().StringP("role", "r", "", "Role (must be 'admin' or 'user')")
	// 	addCmd.Flags().StringP("password", "p", "", "Password")
	// 	addCmd.MarkFlagRequired("user")
	// 	addCmd.MarkFlagRequired("role")
	// 	addCmd.MarkFlagRequired("password")
}
