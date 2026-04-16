package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestCLISmoke(t *testing.T) {
	// Build the CLI
	cmd := exec.Command("go", "build", "-o", "bili-cli-test", "./cmd/bili")
	cmd.Dir = "../.." // Run from project root
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build CLI: %v, output: %s", err, string(out))
	}
	defer os.Remove("../../bili-cli-test")

	// Test help command
	helpCmd := exec.Command("../../bili-cli-test", "--help")
	out, err := helpCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run help command: %v", err)
	}

	if !strings.Contains(string(out), "bili — Bilibili CLI tool") {
		t.Errorf("Expected help output to contain description, got: %s", string(out))
	}

	// Check that subcommands are listed
	expectedCommands := []string{"status", "login", "whoami", "video", "hot", "search"}
	for _, expected := range expectedCommands {
		if !strings.Contains(string(out), expected) {
			t.Errorf("Expected help output to contain command %q", expected)
		}
	}
}
