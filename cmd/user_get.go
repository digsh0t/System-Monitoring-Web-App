package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/wintltr/login-api/models"
)

func init() {
	userGetCmd.PersistentFlags().StringVar(&targetUserArgs.Username, "username", "", "Username of the user you want to see")
	userGetCmd.PersistentFlags().StringVar(&targetUserArgs.Name, "name", "", "Name of the user you want to see")
	userCmd.AddCommand(userGetCmd)
}

var userGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Show user's data",
	Run: func(cmd *cobra.Command, args []string) {

		ok := true

		if targetUserArgs.Username == "" && targetUserArgs.Name == "" {
			fmt.Println("Argument --name or --username required")
			ok = false
		}

		if !ok {
			fmt.Println("Use command `lthmonitor user get --help` for details.")
			os.Exit(1)
		}

		userList, err := models.GetAllUserByUsername(targetUserArgs.Username)
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
		for _, user := range userList {
			fmt.Printf("Username: %s, Name: %s, role: %s\n", user.Username, user.Name, user.Role)
		}
	},
}
