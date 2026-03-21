package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestState_JSON(t *testing.T) {
	created := time.Date(2024, 1, 15, 12, 30, 45, 123456789, time.UTC)
	state := &State{
		Version:     "1.0.2",
		ID:          "test-container",
		Status:      StatusRunning,
		Pid:         1234,
		Bundle:      "/var/lib/membrane/bundles/test",
		Annotations: map[string]string{"key": "value"},
		Created:     created,
	}

	// Marshal
	data, err := json.Marshal(state)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	// Unmarshal
	var loaded State
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if loaded.ID != state.ID {
		t.Errorf("ID = %s, want %s", loaded.ID, state.ID)
	}
	if loaded.Status != state.Status {
		t.Errorf("Status = %s, want %s", loaded.Status, state.Status)
	}
	if loaded.Pid != state.Pid {
		t.Errorf("Pid = %d, want %d", loaded.Pid, state.Pid)
	}
	if !loaded.Created.Equal(state.Created) {
		t.Errorf("Created = %v, want %v", loaded.Created, state.Created)
	}
}

func TestState_StatusMethods(t *testing.T) {
	tests := []struct {
		status     Status
		isRunning  bool
		isStopped  bool
		canStart   bool
		canDelete  bool
	}{
		{StatusCreating, false, false, false, false},
		{StatusCreated, false, false, true, true},
		{StatusRunning, true, false, false, false},
		{StatusStopped, false, true, false, true},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			s := &State{Status: tt.status}
			if s.IsRunning() != tt.isRunning {
				t.Errorf("IsRunning() = %v, want %v", s.IsRunning(), tt.isRunning)
			}
			if s.IsStopped() != tt.isStopped {
				t.Errorf("IsStopped() = %v, want %v", s.IsStopped(), tt.isStopped)
			}
			if s.CanStart() != tt.canStart {
				t.Errorf("CanStart() = %v, want %v", s.CanStart(), tt.canStart)
			}
			if s.CanDelete() != tt.canDelete {
				t.Errorf("CanDelete() = %v, want %v", s.CanDelete(), tt.canDelete)
			}
		})
	}
}

func TestFileStore_SaveLoad(t *testing.T) {
	root := t.TempDir()
	store := NewFileStore(root)

	state := &State{
		Version: "1.0.2",
		ID:      "test-container",
		Status:  StatusCreated,
		Bundle:  "/path/to/bundle",
		Created: time.Now(),
	}

	// Save
	if err := store.Save(state); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Check file exists
	statePath := filepath.Join(root, "test-container", "state.json")
	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		t.Error("state file not created")
	}

	// Load
	loaded, err := store.Load("test-container")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.ID != state.ID {
		t.Errorf("ID = %s, want %s", loaded.ID, state.ID)
	}
	if loaded.Status != state.Status {
		t.Errorf("Status = %s, want %s", loaded.Status, state.Status)
	}
}

func TestFileStore_LoadNotFound(t *testing.T) {
	root := t.TempDir()
	store := NewFileStore(root)

	_, err := store.Load("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent container")
	}
}

func TestFileStore_Delete(t *testing.T) {
	root := t.TempDir()
	store := NewFileStore(root)

	state := &State{
		Version: "1.0.2",
		ID:      "test-container",
		Status:  StatusCreated,
		Bundle:  "/path/to/bundle",
	}

	// Save first
	if err := store.Save(state); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Delete
	if err := store.Delete("test-container"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify deleted
	if store.Exists("test-container") {
		t.Error("container still exists after delete")
	}
}

func TestFileStore_DeleteNotFound(t *testing.T) {
	root := t.TempDir()
	store := NewFileStore(root)

	err := store.Delete("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent container")
	}
}

func TestFileStore_List(t *testing.T) {
	root := t.TempDir()
	store := NewFileStore(root)

	// Create multiple containers
	for i := 0; i < 3; i++ {
		state := &State{
			Version: "1.0.2",
			ID:      fmt.Sprintf("container-%d", i),
			Status:  StatusCreated,
			Bundle:  "/path/to/bundle",
		}
		if err := store.Save(state); err != nil {
			t.Fatalf("Save failed: %v", err)
		}
	}

	// List
	states, err := store.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(states) != 3 {
		t.Errorf("List returned %d containers, want 3", len(states))
	}
}

func TestFileStore_ListEmpty(t *testing.T) {
	root := t.TempDir()
	store := NewFileStore(root)

	states, err := store.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if states != nil && len(states) != 0 {
		t.Errorf("List returned %d containers, want 0", len(states))
	}
}

func TestFileStore_Exists(t *testing.T) {
	root := t.TempDir()
	store := NewFileStore(root)

	if store.Exists("nonexistent") {
		t.Error("Exists returned true for nonexistent container")
	}

	state := &State{
		Version: "1.0.2",
		ID:      "test-container",
		Status:  StatusCreated,
		Bundle:  "/path/to/bundle",
	}
	if err := store.Save(state); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	if !store.Exists("test-container") {
		t.Error("Exists returned false for existing container")
	}
}

func TestFileStore_Create(t *testing.T) {
	root := t.TempDir()
	store := NewFileStore(root)

	state := &State{
		Version: "1.0.2",
		ID:      "test-container",
		Status:  StatusCreated,
		Bundle:  "/path/to/bundle",
	}

	// Create
	if err := store.Create(state); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Create again should fail
	err := store.Create(state)
	if err == nil {
		t.Fatal("expected error for duplicate create")
	}
}

func TestFileStore_DefaultRoot(t *testing.T) {
	store := NewFileStore("")
	if store.Root != "/run/membrane" {
		t.Errorf("default root = %s, want /run/membrane", store.Root)
	}
}
