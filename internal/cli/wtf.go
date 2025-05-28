package cli

import (
	"fmt"
	"runtime"

	"github.com/ephlabs/eph/pkg/version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var wtfCmd = &cobra.Command{
	Use:   "wtf",
	Short: "Display diagnostic information",
	Long:  "Displays diagnostic information to help troubleshoot issues with Eph.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🔍 Eph Diagnostic Information")
		fmt.Println("============================")
		fmt.Printf("Eph Version: %s\n", version.GetVersion())
		fmt.Printf("Go Version: %s\n", version.GoVersion)
		fmt.Printf("Git Commit: %s\n", version.GitCommit)
		fmt.Printf("Built: %s\n", version.BuildDate)
		fmt.Printf("OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
		fmt.Printf("Config file: %s\n", viper.ConfigFileUsed())
		fmt.Printf("Debug mode: %v\n", debug)

		fmt.Println("\n🚧 Advanced diagnostics coming soon!")
		fmt.Println("When in doubt, just eph it!")
	},
}

func init() {
	rootCmd.AddCommand(wtfCmd)
}
