package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ephlabs/eph/pkg/version"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of Eph",
	Long:  "All software has versions. This is Eph's.",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Printf("Eph %s\n", version.GetVersion())
		fmt.Printf("Go Version: %s\n", version.GoVersion)
		fmt.Printf("Git Commit: %s\n", version.GitCommit)
		fmt.Printf("Built: %s\n", version.BuildDate)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
