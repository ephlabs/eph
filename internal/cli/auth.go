package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authentication commands",
	Long:  "Commands to manage authentication with Eph.",
	RunE: func(cmd *cobra.Command, _ []string) error {
		return cmd.Help()
	},
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Log in to Eph",
	Long:  "Authenticate with the Eph service to gain access to your environments.",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Println("🔑 Authentication coming soon!")
		fmt.Println("You'll soon be able to securely log in to Eph here.")
	},
}

func init() {
	rootCmd.AddCommand(authCmd)
	authCmd.AddCommand(loginCmd)
}
