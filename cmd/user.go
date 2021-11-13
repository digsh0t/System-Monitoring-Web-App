package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

type userArgs struct {
	username string
	name     string
	password string
	admin    bool
}

var targetUserArgs userArgs

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
