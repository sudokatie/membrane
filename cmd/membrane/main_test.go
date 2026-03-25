package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestVersionCommand tests the version command output.
func TestVersionCommand(t *testing.T) {
	// Build the binary first
	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "membrane")

	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Dir = "."
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %v\n%s", err, output)
	}

	// Run version command
	var stdout bytes.Buffer
	versionCmd := exec.Command(binaryPath, "version")
	versionCmd.Stdout = &stdout
	if err := versionCmd.Run(); err != nil {
		t.Fatalf("version command failed: %v", err)
	}

	// Parse output
	var version map[string]string
	if err := json.Unmarshal(stdout.Bytes(), &version); err != nil {
		t.Fatalf("parse version output: %v", err)
	}

	if version["ociVersion"] != "1.0.2" {
		t.Errorf("ociVersion = %q, want %q", version["ociVersion"], "1.0.2")
	}
	if version["version"] == "" {
		t.Error("version should not be empty")
	}
}

// TestSpecCommand tests the spec command output.
func TestSpecCommand(t *testing.T) {
	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "membrane")

	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Dir = "."
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %v\n%s", err, output)
	}

	var stdout bytes.Buffer
	specCmd := exec.Command(binaryPath, "spec")
	specCmd.Stdout = &stdout
	if err := specCmd.Run(); err != nil {
		t.Fatalf("spec command failed: %v", err)
	}

	var spec map[string]interface{}
	if err := json.Unmarshal(stdout.Bytes(), &spec); err != nil {
		t.Fatalf("parse spec output: %v", err)
	}

	if spec["ociVersion"] != "1.0.2" {
		t.Errorf("ociVersion = %v, want %q", spec["ociVersion"], "1.0.2")
	}
	if spec["root"] == nil {
		t.Error("spec should have root")
	}
	if spec["process"] == nil {
		t.Error("spec should have process")
	}
}

// TestListCommand tests the list command.
func TestListCommand(t *testing.T) {
	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "membrane")

	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Dir = "."
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %v\n%s", err, output)
	}

	stateRoot := filepath.Join(tmpDir, "state")
	os.MkdirAll(stateRoot, 0755)

	var stdout bytes.Buffer
	listCmd := exec.Command(binaryPath, "--root", stateRoot, "list")
	listCmd.Stdout = &stdout
	if err := listCmd.Run(); err != nil {
		t.Fatalf("list command failed: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "ID") || !strings.Contains(output, "STATUS") {
		t.Errorf("list output should contain headers, got: %s", output)
	}
}

// TestCreateDeleteCommand tests create and delete commands.
func TestCreateDeleteCommand(t *testing.T) {
	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "membrane")

	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Dir = "."
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %v\n%s", err, output)
	}

	// Create bundle
	bundleDir := filepath.Join(tmpDir, "bundle")
	rootfsDir := filepath.Join(bundleDir, "rootfs")
	os.MkdirAll(rootfsDir, 0755)

	configJSON := `{
		"ociVersion": "1.0.2",
		"root": {"path": "rootfs"},
		"process": {"args": ["/bin/sh"], "cwd": "/"}
	}`
	os.WriteFile(filepath.Join(bundleDir, "config.json"), []byte(configJSON), 0644)

	stateRoot := filepath.Join(tmpDir, "state")

	// Create container
	createCmd := exec.Command(binaryPath, "--root", stateRoot, "create", "test-container", bundleDir)
	if output, err := createCmd.CombinedOutput(); err != nil {
		t.Fatalf("create failed: %v\n%s", err, output)
	}

	// State should show created
	var stdout bytes.Buffer
	stateCmd := exec.Command(binaryPath, "--root", stateRoot, "state", "test-container")
	stateCmd.Stdout = &stdout
	if err := stateCmd.Run(); err != nil {
		t.Fatalf("state command failed: %v", err)
	}

	var state map[string]interface{}
	if err := json.Unmarshal(stdout.Bytes(), &state); err != nil {
		t.Fatalf("parse state: %v", err)
	}
	if state["status"] != "created" {
		t.Errorf("status = %v, want created", state["status"])
	}

	// Delete container
	deleteCmd := exec.Command(binaryPath, "--root", stateRoot, "delete", "test-container")
	if output, err := deleteCmd.CombinedOutput(); err != nil {
		t.Fatalf("delete failed: %v\n%s", err, output)
	}

	// State should fail now
	stateCmd2 := exec.Command(binaryPath, "--root", stateRoot, "state", "test-container")
	if err := stateCmd2.Run(); err == nil {
		t.Error("state should fail after delete")
	}
}

// TestInvalidCommands tests error handling for invalid commands.
func TestInvalidCommands(t *testing.T) {
	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "membrane")

	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Dir = "."
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %v\n%s", err, output)
	}

	stateRoot := filepath.Join(tmpDir, "state")

	tests := []struct {
		name string
		args []string
	}{
		{"create without args", []string{"create"}},
		{"create with one arg", []string{"create", "id-only"}},
		{"start without args", []string{"start"}},
		{"state without args", []string{"state"}},
		{"kill without args", []string{"kill"}},
		{"delete without args", []string{"delete"}},
		{"state nonexistent", []string{"state", "nonexistent"}},
		{"delete nonexistent", []string{"delete", "nonexistent"}},
		{"create invalid bundle", []string{"create", "test", "/nonexistent/bundle"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := append([]string{"--root", stateRoot}, tt.args...)
			cmd := exec.Command(binaryPath, args...)
			if err := cmd.Run(); err == nil {
				t.Errorf("%s should fail", tt.name)
			}
		})
	}
}

// TestLogLevel tests log level flag.
func TestLogLevel(t *testing.T) {
	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "membrane")

	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Dir = "."
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %v\n%s", err, output)
	}

	// Should accept different log levels
	levels := []string{"error", "warn", "info", "debug"}
	for _, level := range levels {
		cmd := exec.Command(binaryPath, "--log-level", level, "version")
		if err := cmd.Run(); err != nil {
			t.Errorf("--log-level=%s should work: %v", level, err)
		}
	}
}
