package cli

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCompletionCommand(t *testing.T) {
	shells := []string{"bash", "zsh", "fish", "powershell"}

	for _, shell := range shells {
		t.Run(shell, func(t *testing.T) {
			// Save and restore stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Execute the command
			rootCmd.SetArgs([]string{"completion", shell})
			err := rootCmd.Execute()
			if err != nil {
				t.Errorf("completion %s command failed: %v", shell, err)
			}

			// Restore stdout and capture output
			w.Close()
			os.Stdout = oldStdout

			var buf bytes.Buffer
			_, err = io.Copy(&buf, r)
			require.NoError(t, err)
			output := buf.String()

			// Every shell should output some completion code
			if len(output) < 10 {
				t.Errorf("expected completion output for %s to be non-empty, got: %s", shell, output)
			}

			// Check for shell-specific markers
			var marker string
			switch shell {
			case "bash":
				marker = "bash completion for eph"
			case "zsh":
				marker = "# zsh completion for eph"
			case "fish":
				marker = "# fish completion for eph"
			case "powershell":
				marker = "Register-ArgumentCompleter"
			}

			if !strings.Contains(output, marker) {
				t.Errorf("expected %s completion output to contain '%s', got: %s", shell, marker, output[:100]+"...")
			}
		})
	}
}

func TestCompletionInvalidArgs(t *testing.T) {
	// Test with invalid shell
	rootCmd.SetArgs([]string{"completion", "invalid-shell"})
	err := rootCmd.Execute()

	if err == nil {
		t.Error("expected error for invalid shell, got nil")
	}
}
