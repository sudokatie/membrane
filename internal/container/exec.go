package container

import (
	"fmt"
	"os"
	"syscall"

	"github.com/sudokatie/membrane/pkg/oci"
)

// execProcess executes the container's init process.
// This replaces the current process via execve.
func (m *Manager) execProcess(process *oci.Process) error {
	if process == nil || len(process.Args) == 0 {
		return fmt.Errorf("no process args specified")
	}

	// Set environment
	env := process.Env
	if len(env) == 0 {
		env = []string{
			"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
			"TERM=xterm",
		}
	}

	// Set working directory
	if process.Cwd != "" {
		if err := os.Chdir(process.Cwd); err != nil {
			return fmt.Errorf("chdir to %s: %w", process.Cwd, err)
		}
	}

	// Set user/group (if not root)
	if process.User.GID != 0 {
		if err := syscall.Setgid(int(process.User.GID)); err != nil {
			return fmt.Errorf("setgid: %w", err)
		}
	}
	if len(process.User.AdditionalGids) > 0 {
		gids := make([]int, len(process.User.AdditionalGids))
		for i, gid := range process.User.AdditionalGids {
			gids[i] = int(gid)
		}
		if err := syscall.Setgroups(gids); err != nil {
			return fmt.Errorf("setgroups: %w", err)
		}
	}
	if process.User.UID != 0 {
		if err := syscall.Setuid(int(process.User.UID)); err != nil {
			return fmt.Errorf("setuid: %w", err)
		}
	}

	// Find the executable
	argv0 := process.Args[0]

	// Execute the process
	return syscall.Exec(argv0, process.Args, env)
}

// ExecOptions holds options for executing a command in a container.
type ExecOptions struct {
	// ID is the container ID.
	ID string
	// Args is the command and arguments.
	Args []string
	// Env is additional environment variables.
	Env []string
	// Cwd is the working directory.
	Cwd string
	// User is the user to run as.
	User *oci.User
}

// Exec executes a command in a running container.
func (m *Manager) Exec(opts *ExecOptions) error {
	// Load state
	st, err := m.store.Load(opts.ID)
	if err != nil {
		return err
	}

	// Check state
	if !st.IsRunning() {
		return fmt.Errorf("container is not running")
	}

	// Build process config
	process := &oci.Process{
		Args: opts.Args,
		Env:  opts.Env,
		Cwd:  opts.Cwd,
	}
	if opts.User != nil {
		process.User = *opts.User
	}

	return m.execProcess(process)
}

// closeExtraFDs closes all file descriptors except stdin, stdout, stderr.
func closeExtraFDs() error {
	// Read /proc/self/fd to find open fds
	dir, err := os.Open("/proc/self/fd")
	if err != nil {
		return err
	}
	defer dir.Close()

	fds, err := dir.Readdirnames(-1)
	if err != nil {
		return err
	}

	for _, fd := range fds {
		var fdNum int
		if _, err := fmt.Sscanf(fd, "%d", &fdNum); err != nil {
			continue
		}
		// Keep 0, 1, 2 (stdin, stdout, stderr) and the dir fd
		if fdNum > 2 && fdNum != int(dir.Fd()) {
			syscall.Close(fdNum)
		}
	}

	return nil
}
