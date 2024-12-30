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

		// Get the role and verify it is admin or user
		role, _ := cmd.Flags().GetString("role")
		if role != "admin" && role != "user" {
			fmt.Println("ERR - Role must be 'admin' or 'user'")
			os.Exit(1)
		}

		// Ensure the password is not empty
		password, _ := cmd.Flags().GetString("password")
		if password == "" {
			fmt.Println("ERR - Password cannot be empty")
			os.Exit(1)
		}

		// Ensure the username is not empty
		username, _ := cmd.Flags().GetString("user")
		if username == "" {
			fmt.Println("ERR - Username cannot be empty")
			os.Exit(1)
		}

		user := &models.User{
			Username:     username,
			PasswordHash: auth.GeneratePassword(password),
		}

		if role == "admin" {
			user.Role = types.UserRoleAdmin
		} else {
			user.Role = types.UserRoleUser
		}

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

		options := &database.Options{
			Where: squirrel.Eq{models.USER_TABLE + "." + models.USER_USERNAME: username},
		}

		err = dao.Get(ctx, user, options)
		if err != nil && err != sql.ErrNoRows {
			fmt.Printf("ERR - Failed to look up users: %s", err)
			os.Exit(1)
		}

		if err == nil {
			fmt.Println("ERR - Username already exists")
			os.Exit(1)
		}

		err = dao.CreateUser(ctx, user)
		if err != nil {
			fmt.Printf("ERR - Failed to create user: %s", err)
			os.Exit(1)
		}

		fmt.Printf("User '%s' created\n", username)
	},
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func init() {
	userCmd.AddCommand(addCmd)

	addCmd.Flags().StringP("user", "u", "", "Username")
	addCmd.Flags().StringP("role", "r", "", "Role (must be 'admin' or 'user')")
	addCmd.Flags().StringP("password", "p", "", "Password")
	addCmd.MarkFlagRequired("user")
	addCmd.MarkFlagRequired("role")
	addCmd.MarkFlagRequired("password")
}
