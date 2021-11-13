package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(serviceCmd)
}

var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "Run LTH Monitor service",
	Run: func(cmd *cobra.Command, args []string) {
		runService()
		fmt.Println("Run service")
	},
}
