package cli

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestUpCommand(t *testing.T) {
	runPlaceholderTest(t, "up", []string{}, []string{
		"Environment creation coming soon",
		"What the eph",
	})
}

func TestDownCommand(t *testing.T) {
	runPlaceholderTest(t, "down", []string{}, []string{
		"Environment destruction coming soon",
	})
}

func TestListCommand(t *testing.T) {
	runPlaceholderTest(t, "list", []string{}, []string{
		"Environment listing coming soon",
	})
}

func TestLogsCommand(t *testing.T) {
	runPlaceholderTest(t, "logs", []string{}, []string{
		"Log streaming coming soon",
	})
}

func TestAuthCommand(t *testing.T) {
	// Test auth root command (should show help)
	// Save and restore stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Execute the command
	rootCmd.SetArgs([]string{"auth"})
	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("auth command failed: %v", err)
	}

	// Restore stdout and capture output
	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	expectedPhrases := []string{"Commands to manage authentication", "Usage"}
	for _, phrase := range expectedPhrases {
		if !strings.Contains(output, phrase) {
			t.Errorf("expected auth output to contain '%s', got: %s", phrase, output)
		}
	}
}

func TestAuthLoginCommand(t *testing.T) {
	runPlaceholderTest(t, "auth", []string{"login"}, []string{
		"Authentication coming soon",
	})
}

// runPlaceholderTest executes a command with args and checks if the output contains expected phrases
func runPlaceholderTest(t *testing.T, command string, args []string, expectedPhrases []string) {
	// Save and restore stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Build command args
	cmdArgs := []string{command}
	cmdArgs = append(cmdArgs, args...)
	rootCmd.SetArgs(cmdArgs)

	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("%s command failed: %v", command, err)
	}

	// Restore stdout and capture output
	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	for _, phrase := range expectedPhrases {
		if !strings.Contains(output, phrase) {
			t.Errorf("expected %s output to contain '%s', got: %s", command, phrase, output)
		}
	}
}
