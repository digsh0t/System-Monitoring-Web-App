package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/wintltr/login-api/models"
)

func init() {
	userDeleteCmd.PersistentFlags().StringVar(&targetUserArgs.Username, "username", "", "Username of the user you want to delete")
	userDeleteCmd.PersistentFlags().StringVar(&targetUserArgs.Name, "name", "", "Name of the user you want to delete")
	userCmd.AddCommand(userDeleteCmd)
}

var userDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Remove existing user",
	Run: func(cmd *cobra.Command, args []string) {

		ok := true

		if targetUserArgs.Username == "" && targetUserArgs.Name == "" {
			fmt.Println("Argument --username or --name required")
			ok = false
		}

		if !ok {
			fmt.Println("Use command `lthmonitor user delete --help` for details.")
			os.Exit(1)
		}

		err := models.DeleteUserWithUsernameFromDB(targetUserArgs.Username)
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}

		fmt.Printf("User %s deleted!\n", targetUserArgs.Username)
	},
}
