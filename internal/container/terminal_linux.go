//go:build linux

package container

import (
	"fmt"
	"os"
	"unsafe"

	"golang.org/x/sys/unix"
)

// Terminal represents a pseudo-terminal.
type Terminal struct {
	// Master is the master side of the PTY.
	Master *os.File
	// Slave is the slave side of the PTY.
	Slave *os.File
	// SlavePath is the path to the slave device.
	SlavePath string
}

// NewTerminal creates a new pseudo-terminal.
func NewTerminal() (*Terminal, error) {
	// Open the PTY master
	master, err := os.OpenFile("/dev/ptmx", os.O_RDWR|unix.O_NOCTTY|unix.O_CLOEXEC, 0)
	if err != nil {
		return nil, fmt.Errorf("open /dev/ptmx: %w", err)
	}

	// Grant access to the slave
	if err := grantpt(master); err != nil {
		master.Close()
		return nil, fmt.Errorf("grantpt: %w", err)
	}

	// Unlock the slave
	if err := unlockpt(master); err != nil {
		master.Close()
		return nil, fmt.Errorf("unlockpt: %w", err)
	}

	// Get the slave path
	slavePath, err := ptsname(master)
	if err != nil {
		master.Close()
		return nil, fmt.Errorf("ptsname: %w", err)
	}

	// Open the slave
	slave, err := os.OpenFile(slavePath, os.O_RDWR|unix.O_NOCTTY, 0)
	if err != nil {
		master.Close()
		return nil, fmt.Errorf("open slave: %w", err)
	}

	return &Terminal{
		Master:    master,
		Slave:     slave,
		SlavePath: slavePath,
	}, nil
}

// Close closes both sides of the terminal.
func (t *Terminal) Close() error {
	var errs []error
	if t.Master != nil {
		if err := t.Master.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if t.Slave != nil {
		if err := t.Slave.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return errs[0]
	}
	return nil
}

// SetupChildTerminal sets up the terminal for the child process.
// This should be called in the child after fork.
func (t *Terminal) SetupChildTerminal() error {
	// Create a new session
	if _, err := unix.Setsid(); err != nil {
		return fmt.Errorf("setsid: %w", err)
	}

	// Set the slave as the controlling terminal
	if err := unix.IoctlSetInt(int(t.Slave.Fd()), unix.TIOCSCTTY, 0); err != nil {
		return fmt.Errorf("TIOCSCTTY: %w", err)
	}

	// Duplicate the slave to stdin, stdout, stderr
	for _, fd := range []int{0, 1, 2} {
		if err := unix.Dup2(int(t.Slave.Fd()), fd); err != nil {
			return fmt.Errorf("dup2 to fd %d: %w", fd, err)
		}
	}

	// Close the original slave fd if it's not 0, 1, or 2
	slaveFd := int(t.Slave.Fd())
	if slaveFd > 2 {
		unix.Close(slaveFd)
	}

	return nil
}

// grantpt grants access to the slave pseudo-terminal.
func grantpt(master *os.File) error {
	// On modern Linux, this is a no-op as devpts handles permissions
	return nil
}

// unlockpt unlocks the slave pseudo-terminal.
func unlockpt(master *os.File) error {
	var unlock int32 = 0
	_, _, errno := unix.Syscall(unix.SYS_IOCTL, master.Fd(), unix.TIOCSPTLCK, uintptr(unsafe.Pointer(&unlock)))
	if errno != 0 {
		return errno
	}
	return nil
}

// ptsname returns the name of the slave pseudo-terminal.
func ptsname(master *os.File) (string, error) {
	var ptyno uint32
	_, _, errno := unix.Syscall(unix.SYS_IOCTL, master.Fd(), unix.TIOCGPTN, uintptr(unsafe.Pointer(&ptyno)))
	if errno != 0 {
		return "", errno
	}
	return fmt.Sprintf("/dev/pts/%d", ptyno), nil
}

// SetWinsize sets the terminal window size.
func SetWinsize(fd uintptr, rows, cols uint16) error {
	ws := unix.Winsize{
		Row: rows,
		Col: cols,
	}
	_, _, errno := unix.Syscall(unix.SYS_IOCTL, fd, unix.TIOCSWINSZ, uintptr(unsafe.Pointer(&ws)))
	if errno != 0 {
		return errno
	}
	return nil
}
