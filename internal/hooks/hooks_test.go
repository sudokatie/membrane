package hooks

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sudokatie/membrane/internal/state"
	"github.com/sudokatie/membrane/pkg/oci"
)

func TestRun(t *testing.T) {
	// Create a test script
	dir := t.TempDir()
	scriptPath := filepath.Join(dir, "hook.sh")
	err := os.WriteFile(scriptPath, []byte("#!/bin/sh\nexit 0\n"), 0755)
	if err != nil {
		t.Fatalf("create script: %v", err)
	}

	hooks := []oci.Hook{
		{
			Path: scriptPath,
			Args: []string{scriptPath},
		},
	}

	hookState := &HookState{
		OCIVersion: "1.0.2",
		ID:         "test-container",
		Status:     state.StatusCreated,
		Bundle:     "/test/bundle",
	}

	err = Run(hooks, hookState)
	if err != nil {
		t.Errorf("Run failed: %v", err)
	}
}

func TestRunTimeout(t *testing.T) {
	// Create a script that sleeps
	dir := t.TempDir()
	scriptPath := filepath.Join(dir, "sleep.sh")
	err := os.WriteFile(scriptPath, []byte("#!/bin/sh\nsleep 10\n"), 0755)
	if err != nil {
		t.Fatalf("create script: %v", err)
	}

	timeout := 1
	hooks := []oci.Hook{
		{
			Path:    scriptPath,
			Args:    []string{scriptPath},
			Timeout: &timeout,
		},
	}

	hookState := &HookState{
		OCIVersion: "1.0.2",
		ID:         "test-container",
		Status:     state.StatusCreated,
		Bundle:     "/test/bundle",
	}

	err = Run(hooks, hookState)
	if err == nil {
		t.Error("expected timeout error")
	}
}

func TestRunFailure(t *testing.T) {
	// Create a script that fails
	dir := t.TempDir()
	scriptPath := filepath.Join(dir, "fail.sh")
	err := os.WriteFile(scriptPath, []byte("#!/bin/sh\nexit 1\n"), 0755)
	if err != nil {
		t.Fatalf("create script: %v", err)
	}

	hooks := []oci.Hook{
		{
			Path: scriptPath,
			Args: []string{scriptPath},
		},
	}

	hookState := &HookState{
		OCIVersion: "1.0.2",
		ID:         "test-container",
		Status:     state.StatusCreated,
		Bundle:     "/test/bundle",
	}

	err = Run(hooks, hookState)
	if err == nil {
		t.Error("expected hook failure error")
	}
}

func TestRunNilHooks(t *testing.T) {
	hookState := &HookState{
		OCIVersion: "1.0.2",
		ID:         "test-container",
		Status:     state.StatusCreated,
		Bundle:     "/test/bundle",
	}

	// Should not error on nil hooks
	if err := RunPrestart(nil, hookState); err != nil {
		t.Errorf("RunPrestart with nil failed: %v", err)
	}
	if err := RunCreateRuntime(nil, hookState); err != nil {
		t.Errorf("RunCreateRuntime with nil failed: %v", err)
	}
	if err := RunCreateContainer(nil, hookState); err != nil {
		t.Errorf("RunCreateContainer with nil failed: %v", err)
	}
	if err := RunStartContainer(nil, hookState); err != nil {
		t.Errorf("RunStartContainer with nil failed: %v", err)
	}
	if err := RunPoststart(nil, hookState); err != nil {
		t.Errorf("RunPoststart with nil failed: %v", err)
	}
	if err := RunPoststop(nil, hookState); err != nil {
		t.Errorf("RunPoststop with nil failed: %v", err)
	}
}

func TestRunEmptyHooks(t *testing.T) {
	hookState := &HookState{
		OCIVersion: "1.0.2",
		ID:         "test-container",
		Status:     state.StatusCreated,
		Bundle:     "/test/bundle",
	}

	hooks := &oci.Hooks{}

	// Should not error on empty hooks
	if err := RunPrestart(hooks, hookState); err != nil {
		t.Errorf("RunPrestart with empty failed: %v", err)
	}
	if err := RunCreateRuntime(hooks, hookState); err != nil {
		t.Errorf("RunCreateRuntime with empty failed: %v", err)
	}
}
