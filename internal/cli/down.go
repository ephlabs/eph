package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var downCmd = &cobra.Command{
	Use:   "down",
	Short: "Destroy an ephemeral environment",
	Long:  "Tear down an ephemeral environment when you're done with it.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("💥 Environment destruction coming soon!")
		fmt.Println("Soon you'll be able to tear down environments with ease.")
	},
}

func init() {
	rootCmd.AddCommand(downCmd)
}
