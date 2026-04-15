package cmd

import (
	"testing"
)

func TestWorkspaceFlag(t *testing.T) {
	// Reset for test
	workspacePath = ""

	rootCmd.SetArgs([]string{"--workspace", "/tmp/test-ws", "config", "show"})
	if err := rootCmd.Execute(); err != nil {
		// config show may fail, that's ok — we just check the flag was parsed
	}

	if workspacePath != "/tmp/test-ws" {
		t.Fatalf("workspacePath: got %q, want /tmp/test-ws", workspacePath)
	}
}
