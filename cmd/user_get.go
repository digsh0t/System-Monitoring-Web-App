package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	userGetCmd.PersistentFlags().StringVar(&targetUserArgs.username, "username", "", "Username of the user you want to see")
	userGetCmd.PersistentFlags().StringVar(&targetUserArgs.name, "name", "", "Name of the user you want to see")
	userCmd.AddCommand(userGetCmd)
}

var userGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Show user's data",
	Run: func(cmd *cobra.Command, args []string) {

		ok := true

		if targetUserArgs.username == "" && targetUserArgs.name == "" {
			fmt.Println("Argument --name or --username required")
			ok = false
		}

		if !ok {
			fmt.Println("Use command `lthmonitor user get --help` for details.")
			os.Exit(1)
		}

		fmt.Println("Loading user by name from DB (REPLACE THIS)")

		fmt.Println("User info here (REPLACE THIS)")
	},
}
