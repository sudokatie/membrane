package container

import (
	"fmt"
	"os"
	"strconv"
	"syscall"

	"github.com/sudokatie/membrane/internal/log"
	"github.com/sudokatie/membrane/internal/state"
	"github.com/sudokatie/membrane/pkg/oci"
)

// KillOptions holds options for killing a container.
type KillOptions struct {
	// ID is the container ID.
	ID string
	// Signal is the signal to send (default: SIGTERM).
	Signal string
	// All sends the signal to all processes in the container.
	All bool
}

// Kill sends a signal to the container's init process.
func (m *Manager) Kill(opts *KillOptions) error {
	logger := log.WithField("container", opts.ID)

	// Load state
	st, err := m.store.Load(opts.ID)
	if err != nil {
		return err
	}

	// Check state
	if !st.IsRunning() {
		return fmt.Errorf("container is not running")
	}

	// Parse signal
	sig := syscall.SIGTERM
	if opts.Signal != "" {
		if s, ok := oci.Signals[opts.Signal]; ok {
			sig = s
		} else {
			// Try parsing as number
			if n, err := strconv.Atoi(opts.Signal); err == nil {
				sig = syscall.Signal(n)
			} else {
				return fmt.Errorf("unknown signal: %s", opts.Signal)
			}
		}
	}

	logger.WithField("signal", sig).Debug("sending signal to container")

	// Find process
	proc, err := os.FindProcess(st.Pid)
	if err != nil {
		return fmt.Errorf("find process: %w", err)
	}

	// Send signal
	if err := proc.Signal(sig); err != nil {
		// Process may have already exited
		if err == os.ErrProcessDone {
			logger.Debug("process already exited")
			// Update state to stopped
			st.Status = state.StatusStopped
			st.Pid = 0
			m.store.Save(st)
			return nil
		}
		return fmt.Errorf("send signal: %w", err)
	}

	logger.Info("signal sent to container")
	return nil
}

// Wait waits for a container to exit and returns the exit code.
func (m *Manager) Wait(id string) (int, error) {
	logger := log.WithField("container", id)
	logger.Debug("waiting for container to exit")

	// Load state
	st, err := m.store.Load(id)
	if err != nil {
		return -1, err
	}

	// If already stopped, return cached exit code
	if st.IsStopped() {
		if st.ExitCode != nil {
			return *st.ExitCode, nil
		}
		return 0, nil
	}

	if !st.IsRunning() {
		return 0, nil // Not running, not stopped - assume success
	}

	// Wait for the process
	proc, err := os.FindProcess(st.Pid)
	if err != nil {
		return -1, fmt.Errorf("find process: %w", err)
	}

	ps, err := proc.Wait()
	if err != nil {
		// Process may have been reaped by someone else
		logger.Debugf("wait failed (process may have been reaped): %v", err)
		st.Status = state.StatusStopped
		st.Pid = 0
		exitCode := -1
		st.ExitCode = &exitCode
		m.store.Save(st)
		return -1, nil
	}

	// Get exit code
	exitCode := ps.ExitCode()
	logger.WithField("exitCode", exitCode).Debug("container exited")

	// Update state
	st.Status = state.StatusStopped
	st.Pid = 0
	st.ExitCode = &exitCode
	if err := m.store.Save(st); err != nil {
		return exitCode, fmt.Errorf("save state: %w", err)
	}

	return exitCode, nil
}

// State returns the current state of a container.
func (m *Manager) State(id string) (*state.State, error) {
	st, err := m.store.Load(id)
	if err != nil {
		return nil, err
	}

	// Check if process is still running
	if st.IsRunning() && st.Pid > 0 {
		proc, err := os.FindProcess(st.Pid)
		if err != nil {
			// Process doesn't exist, update state
			st.Status = state.StatusStopped
			st.Pid = 0
			m.store.Save(st)
		} else {
			// Check if process is alive by sending signal 0
			if err := proc.Signal(syscall.Signal(0)); err != nil {
				st.Status = state.StatusStopped
				st.Pid = 0
				m.store.Save(st)
			}
		}
	}

	return st, nil
}
