package cli

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestVersionCommand(t *testing.T) {
	// Save and restore stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Execute the command
	rootCmd.SetArgs([]string{"version"})
	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("version command failed: %v", err)
	}

	// Restore stdout and capture output
	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	expectedPhrases := []string{"Eph", "Go Version", "Git Commit", "Built"}
	for _, phrase := range expectedPhrases {
		if !strings.Contains(output, phrase) {
			t.Errorf("expected version output to contain '%s', got: %s", phrase, output)
		}
	}
}
