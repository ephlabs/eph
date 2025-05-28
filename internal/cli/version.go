package cli

import (
	"fmt"

	"github.com/ephlabs/eph/pkg/version"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of Eph",
	Long:  "All software has versions. This is Eph's.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Eph %s\n", version.GetVersion())
		fmt.Printf("Go Version: %s\n", version.GoVersion)
		fmt.Printf("Git Commit: %s\n", version.GitCommit)
		fmt.Printf("Built: %s\n", version.BuildDate)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
