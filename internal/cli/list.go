package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List ephemeral environments",
	Long:  "List all ephemeral environments managed by Eph.",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Println("📋 Environment listing coming soon!")
		fmt.Println("You'll soon be able to see all your ephemeral environments here.")
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
