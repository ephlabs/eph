package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	debug   bool
)

var rootCmd = &cobra.Command{
	Use:   "eph",
	Short: "Ephemeral environment controller - What the eph?",
	Long: `Eph creates and manages ephemeral environments for your pull requests.
	
When you need a preview environment, just eph it!

Eph automatically spins up isolated environments when you label your PRs,
giving every feature branch its own playground. No more "works on my machine"!`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Persistent flags available to all commands
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: ./eph.yaml)")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "enable debug logging")

	// Configure completion command
	rootCmd.CompletionOptions.DisableDefaultCmd = false
	rootCmd.SetHelpTemplate(helpTemplate())

	// Register completion command
	rootCmd.AddCommand(completionCmd)
}

// initConfig reads in config file and ENV variables if set
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag
		viper.SetConfigFile(cfgFile)
	} else {
		// Search for config in current directory with name "eph.yaml"
		viper.AddConfigPath(".")
		viper.SetConfigName("eph")
		viper.SetConfigType("yaml")
	}

	// Read in environment variables that match
	viper.AutomaticEnv()
	viper.SetEnvPrefix("EPH")

	// If a config file is found, read it in
	if err := viper.ReadInConfig(); err == nil && debug {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func helpTemplate() string {
	return `{{with (or .Long .Short)}}{{. | trimTrailingWhitespaces}}

{{end}}{{if or .Runnable .HasSubCommands}}{{.UsageString}}{{end}}

When in doubt, just eph it! 🚀
`
}

// completionCmd represents the completion command
var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion scripts",
	Long: `Generate shell completion scripts for eph CLI.
To load completions:

Bash:
  $ source <(eph completion bash)

Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it. You can execute the following once:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc
  
  # To load completions for each session, execute once:
  $ eph completion zsh > "${fpath[1]}/_eph"
  
  # You will need to start a new shell for this setup to take effect.

Fish:
  $ eph completion fish | source
  
  # To load completions for each session, execute once:
  $ eph completion fish > ~/.config/fish/completions/eph.fish

PowerShell:
  PS> eph completion powershell | Out-String | Invoke-Expression
  
  # To load completions for every new session, run:
  PS> eph completion powershell > eph.ps1
  # and source this file from your PowerShell profile.
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	Run: func(cmd *cobra.Command, args []string) {
		switch args[0] {
		case "bash":
			cmd.Root().GenBashCompletion(os.Stdout)
		case "zsh":
			cmd.Root().GenZshCompletion(os.Stdout)
		case "fish":
			cmd.Root().GenFishCompletion(os.Stdout, true)
		case "powershell":
			cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
		}
	},
}
