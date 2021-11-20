package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/wintltr/login-api/models"
)

func init() {
	userAddCmd.PersistentFlags().StringVar(&targetUserArgs.Username, "username", "", "New user login")
	userAddCmd.PersistentFlags().StringVar(&targetUserArgs.Name, "name", "", "New user name")
	userAddCmd.PersistentFlags().StringVar(&targetUserArgs.Password, "password", "", "New user password")
	userAddCmd.PersistentFlags().StringVar(&targetUserArgs.Role, "role", "user", "Mark new user as admin")
	userCmd.AddCommand(userAddCmd)
}

var userAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add new user",
	Run: func(cmd *cobra.Command, args []string) {

		ok := true
		if targetUserArgs.Name == "" {
			fmt.Println("Argument --name required")
			ok = false
		}
		if targetUserArgs.Username == "" {
			fmt.Println("Argument --username required")
			ok = false
		}

		if targetUserArgs.Password == "" {
			fmt.Println("Argument --password required")
			ok = false
		}

		if !ok {
			fmt.Println("Use command `lthmonitor user add --help` for details.")
			os.Exit(1)
		}

		_, err := models.AddWebAppUser(targetUserArgs)
		if err != nil {
			log.Println(err)
		}

		fmt.Printf("User %s <%s> added!\n", targetUserArgs.Username, targetUserArgs.Name)
	},
}
