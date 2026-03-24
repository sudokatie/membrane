//go:build !linux

package container

import (
	"fmt"
	"os"
)

// Terminal represents a pseudo-terminal stub.
type Terminal struct {
	Master    *os.File
	Slave     *os.File
	SlavePath string
}

// NewTerminal is a stub for non-Linux systems.
func NewTerminal() (*Terminal, error) {
	return nil, fmt.Errorf("terminals not supported on this platform")
}

// Close is a stub for non-Linux systems.
func (t *Terminal) Close() error {
	return nil
}

// SetupChildTerminal is a stub for non-Linux systems.
func (t *Terminal) SetupChildTerminal() error {
	return fmt.Errorf("terminals not supported on this platform")
}

// SetWinsize is a stub for non-Linux systems.
func SetWinsize(fd uintptr, rows, cols uint16) error {
	return fmt.Errorf("terminals not supported on this platform")
}
