package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	userDeleteCmd.PersistentFlags().StringVar(&targetUserArgs.username, "username", "", "Username of the user you want to delete")
	userDeleteCmd.PersistentFlags().StringVar(&targetUserArgs.name, "name", "", "Name of the user you want to delete")
	userCmd.AddCommand(userDeleteCmd)
}

var userDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Remove existing user",
	Run: func(cmd *cobra.Command, args []string) {

		ok := true

		if targetUserArgs.username == "" && targetUserArgs.name == "" {
			fmt.Println("Argument --username or --name required")
			ok = false
		}

		if !ok {
			fmt.Println("Use command `lthmonitor user delete --help` for details.")
			os.Exit(1)
		}

		fmt.Println("Getting user from DB (REPLACE THIS)")
		fmt.Println("Removing user from DB (REPLACE THIS)")

		fmt.Printf("User %s <%s> deleted!\n", "REPLACE THIS", "REPLACE THIS")
	},
}
