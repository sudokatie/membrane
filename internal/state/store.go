package state

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

var (
	// ErrNotFound is returned when a container is not found.
	ErrNotFound = errors.New("container not found")
	// ErrExists is returned when a container already exists.
	ErrExists = errors.New("container already exists")
	// ErrLocked is returned when the state file is locked.
	ErrLocked = errors.New("state file locked")
)

// Store is the interface for container state storage.
type Store interface {
	// Save saves the state for a container.
	Save(state *State) error
	// Load loads the state for a container.
	Load(id string) (*State, error)
	// Delete removes the state for a container.
	Delete(id string) error
	// List returns all container states.
	List() ([]*State, error)
	// Exists returns true if a container exists.
	Exists(id string) bool
}

// FileStore stores container state in the filesystem.
type FileStore struct {
	// Root is the root directory for state storage.
	// Default: /run/membrane
	Root string
}

// NewFileStore creates a new file-based state store.
func NewFileStore(root string) *FileStore {
	if root == "" {
		root = "/run/membrane"
	}
	return &FileStore{Root: root}
}

// containerDir returns the directory for a container's state.
func (s *FileStore) containerDir(id string) string {
	return filepath.Join(s.Root, id)
}

// statePath returns the path to a container's state file.
func (s *FileStore) statePath(id string) string {
	return filepath.Join(s.containerDir(id), "state.json")
}

// lockPath returns the path to a container's lock file.
func (s *FileStore) lockPath(id string) string {
	return filepath.Join(s.containerDir(id), "lock")
}

// Save saves the state for a container.
func (s *FileStore) Save(state *State) error {
	dir := s.containerDir(state.ID)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("create state dir: %w", err)
	}

	// Acquire lock
	lockFile, err := s.acquireLock(state.ID)
	if err != nil {
		return err
	}
	defer s.releaseLock(lockFile)

	// Write state
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal state: %w", err)
	}

	if err := os.WriteFile(s.statePath(state.ID), data, 0600); err != nil {
		return fmt.Errorf("write state: %w", err)
	}

	return nil
}

// Load loads the state for a container.
func (s *FileStore) Load(id string) (*State, error) {
	path := s.statePath(id)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%w: %s", ErrNotFound, id)
		}
		return nil, fmt.Errorf("read state: %w", err)
	}

	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("unmarshal state: %w", err)
	}

	return &state, nil
}

// Delete removes the state for a container.
func (s *FileStore) Delete(id string) error {
	dir := s.containerDir(id)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("%w: %s", ErrNotFound, id)
	}

	// Acquire lock
	lockFile, err := s.acquireLock(id)
	if err != nil {
		return err
	}
	defer s.releaseLock(lockFile)

	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf("remove state dir: %w", err)
	}

	return nil
}

// List returns all container states.
func (s *FileStore) List() ([]*State, error) {
	entries, err := os.ReadDir(s.Root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read state root: %w", err)
	}

	var states []*State
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		state, err := s.Load(entry.Name())
		if err != nil {
			// Skip containers with invalid state
			continue
		}
		states = append(states, state)
	}

	return states, nil
}

// Exists returns true if a container exists.
func (s *FileStore) Exists(id string) bool {
	_, err := os.Stat(s.statePath(id))
	return err == nil
}

// acquireLock acquires an exclusive lock for a container.
func (s *FileStore) acquireLock(id string) (*os.File, error) {
	lockPath := s.lockPath(id)
	
	// Create lock file
	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, fmt.Errorf("create lock file: %w", err)
	}

	// Try to acquire exclusive lock (non-blocking)
	err = syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if err != nil {
		f.Close()
		if err == syscall.EWOULDBLOCK {
			return nil, ErrLocked
		}
		return nil, fmt.Errorf("acquire lock: %w", err)
	}

	return f, nil
}

// releaseLock releases a lock.
func (s *FileStore) releaseLock(f *os.File) {
	if f == nil {
		return
	}
	syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
	f.Close()
}

// Create creates a new container state.
// Returns ErrExists if the container already exists.
func (s *FileStore) Create(state *State) error {
	if s.Exists(state.ID) {
		return fmt.Errorf("%w: %s", ErrExists, state.ID)
	}
	return s.Save(state)
}
