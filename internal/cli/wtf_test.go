package cli

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWtfCommand(t *testing.T) {
	// Save and restore stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Execute the command
	rootCmd.SetArgs([]string{"wtf"})
	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("wtf command failed: %v", err)
	}

	// Restore stdout and capture output
	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	require.NoError(t, err)
	output := buf.String()

	expectedPhrases := []string{
		"Eph Diagnostic Information",
		"Eph Version",
		"Go Version",
		"Git Commit",
		"Built",
		"OS/Arch",
		"Config file",
		"Debug mode",
		"Advanced diagnostics coming soon",
	}

	for _, phrase := range expectedPhrases {
		if !strings.Contains(output, phrase) {
			t.Errorf("expected wtf output to contain '%s', got: %s", phrase, output)
		}
	}
}
