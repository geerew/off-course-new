package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/geerew/off-course/dao"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/appFs"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List users",
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

		dao := dao.NewDAO(dbManager.DataDb)

		users := []*models.User{}
		err = dao.List(ctx, &users, nil)
		if err != nil {
			errorMessage("Failed to list users: %s", err)
			os.Exit(1)
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Username", "Role", "Display Name"})
		table.SetBorder(false)

		for _, user := range users {
			table.Append([]string{user.Username, user.Role.String(), user.DisplayName})
		}
		table.Render()
	},
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func init() {
	userCmd.AddCommand(listCmd)
}
