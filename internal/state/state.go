// Package state manages container state persistence.
package state

import (
	"encoding/json"
	"time"
)

// Status is the container status.
type Status string

// Container statuses per OCI runtime spec.
const (
	// StatusCreating indicates the container is being created.
	StatusCreating Status = "creating"
	// StatusCreated indicates the container has been created but not started.
	StatusCreated Status = "created"
	// StatusRunning indicates the container is running.
	StatusRunning Status = "running"
	// StatusStopped indicates the container has exited.
	StatusStopped Status = "stopped"
)

// State represents the state of a container.
// This matches the OCI runtime state specification.
type State struct {
	// Version is the OCI spec version.
	Version string `json:"ociVersion"`
	// ID is the container's unique identifier.
	ID string `json:"id"`
	// Status is the container's status.
	Status Status `json:"status"`
	// Pid is the process ID of the container's init process.
	// 0 if the container is not running.
	Pid int `json:"pid,omitempty"`
	// Bundle is the absolute path to the container's bundle directory.
	Bundle string `json:"bundle"`
	// Annotations are key-value pairs associated with the container.
	Annotations map[string]string `json:"annotations,omitempty"`
	// Created is the creation timestamp.
	Created time.Time `json:"created,omitempty"`
}

// MarshalJSON implements json.Marshaler.
func (s *State) MarshalJSON() ([]byte, error) {
	type Alias State
	return json.Marshal(&struct {
		Created string `json:"created,omitempty"`
		*Alias
	}{
		Created: s.Created.Format(time.RFC3339Nano),
		Alias:   (*Alias)(s),
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (s *State) UnmarshalJSON(data []byte) error {
	type Alias State
	aux := &struct {
		Created string `json:"created,omitempty"`
		*Alias
	}{
		Alias: (*Alias)(s),
	}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}
	if aux.Created != "" {
		t, err := time.Parse(time.RFC3339Nano, aux.Created)
		if err != nil {
			return err
		}
		s.Created = t
	}
	return nil
}

// IsRunning returns true if the container is in a running state.
func (s *State) IsRunning() bool {
	return s.Status == StatusRunning
}

// IsStopped returns true if the container has stopped.
func (s *State) IsStopped() bool {
	return s.Status == StatusStopped
}

// CanStart returns true if the container can be started.
func (s *State) CanStart() bool {
	return s.Status == StatusCreated
}

// CanDelete returns true if the container can be deleted.
func (s *State) CanDelete() bool {
	return s.Status == StatusCreated || s.Status == StatusStopped
}
