package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Create an ephemeral environment",
	Long:  "Spin up a new ephemeral environment for your pull request.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🚀 Environment creation coming soon!")
		fmt.Println("When ready, this will create environments that make you say 'What the eph?'")
	},
}

func init() {
	rootCmd.AddCommand(upCmd)
}
