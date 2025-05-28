//go:build integration
// +build integration

package cli

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestCLIIntegration(t *testing.T) {
	// Build the CLI
	buildCmd := exec.Command("go", "build", "-o", "eph-test", "../../cmd/eph")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("failed to build CLI: %v", err)
	}
	defer os.Remove("eph-test")

	tests := []struct {
		name         string
		args         []string
		expectOutput string
	}{
		{"help", []string{"--help"}, "Eph creates and manages"},
		{"version", []string{"version"}, "Eph"},
		{"wtf", []string{"wtf"}, "diagnostic"},
		{"up", []string{"up"}, "Environment creation coming soon"},
		{"down", []string{"down"}, "Environment destruction coming soon"},
		{"list", []string{"list"}, "Environment listing coming soon"},
		{"logs", []string{"logs"}, "Log streaming coming soon"},
		{"auth", []string{"auth", "login"}, "Authentication coming soon"},
		{"completion", []string{"completion", "bash"}, "# bash completion for eph"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command("./eph-test", tt.args...)
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("command failed: %v", err)
			}

			if !strings.Contains(string(output), tt.expectOutput) {
				t.Errorf("expected output to contain '%s', got: %s", tt.expectOutput, output)
			}
		})
	}
}
