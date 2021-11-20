package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/wintltr/login-api/models"
)

var targetUserArgs models.User

func init() {
	rootCmd.AddCommand(userCmd)
}

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "Manage users",
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
		os.Exit(0)
	},
}
