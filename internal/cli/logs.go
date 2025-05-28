package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Stream logs from an environment",
	Long:  "Stream logs from an ephemeral environment to debug issues.",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Println("📜 Log streaming coming soon!")
		fmt.Println("Soon you'll be able to see what the eph is happening in your environments.")
	},
}

func init() {
	rootCmd.AddCommand(logsCmd)
}
