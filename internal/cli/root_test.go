package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestRootCommand(t *testing.T) {
	// Create a new command for testing
	cmd := &cobra.Command{Use: "root"}

	// Redirect output
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	// Test help output
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"--help"})

	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	output := buf.String()
	expectedPhrases := []string{"Eph creates and manages", "preview environment", "just eph it"}
	for _, phrase := range expectedPhrases {
		if !strings.Contains(output, phrase) {
			t.Errorf("Expected output to contain '%s', got: %s", phrase, output)
		}
	}
}

func TestGlobalFlags(t *testing.T) {
	// Test if the global flags are properly defined
	flagSet := rootCmd.PersistentFlags()

	// Test config flag
	configFlag := flagSet.Lookup("config")
	if configFlag == nil {
		t.Error("Expected 'config' flag to be defined")
	} else if configFlag.Usage != "config file (default: ./eph.yaml)" {
		t.Errorf("Expected 'config' flag usage to be 'config file (default: ./eph.yaml)', got: %s", configFlag.Usage)
	}

	// Test debug flag
	debugFlag := flagSet.Lookup("debug")
	if debugFlag == nil {
		t.Error("Expected 'debug' flag to be defined")
	} else if debugFlag.Usage != "enable debug logging" {
		t.Errorf("Expected 'debug' flag usage to be 'enable debug logging', got: %s", debugFlag.Usage)
	}
}
