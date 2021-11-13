package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	userAddCmd.PersistentFlags().StringVar(&targetUserArgs.username, "username", "", "New user login")
	userAddCmd.PersistentFlags().StringVar(&targetUserArgs.name, "name", "", "New user name")
	userAddCmd.PersistentFlags().StringVar(&targetUserArgs.password, "password", "", "New user password")
	userAddCmd.PersistentFlags().BoolVar(&targetUserArgs.admin, "admin", false, "Mark new user as admin")
	userCmd.AddCommand(userAddCmd)
}

var userAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add new user",
	Run: func(cmd *cobra.Command, args []string) {

		ok := true
		if targetUserArgs.name == "" {
			fmt.Println("Argument --name required")
			ok = false
		}
		if targetUserArgs.username == "" {
			fmt.Println("Argument --username required")
			ok = false
		}

		if targetUserArgs.password == "" {
			fmt.Println("Argument --password required")
			ok = false
		}

		if !ok {
			fmt.Println("Use command `lthmonitor user add --help` for details.")
			os.Exit(1)
		}

		fmt.Println("Adding new user (REPLACE THIS IN PRODUCT MODE")

		fmt.Printf("User %s <%s> added!\n", targetUserArgs.username, targetUserArgs.name)
	},
}
