// Package hooks implements OCI runtime lifecycle hooks.
package hooks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"

	"github.com/sudokatie/membrane/internal/log"
	"github.com/sudokatie/membrane/internal/state"
	"github.com/sudokatie/membrane/pkg/oci"
)

// DefaultTimeout is the default hook timeout in seconds.
const DefaultTimeout = 30

// HookState is the state passed to hooks via stdin.
type HookState struct {
	// OCIVersion is the OCI spec version.
	OCIVersion string `json:"ociVersion"`
	// ID is the container ID.
	ID string `json:"id"`
	// Status is the container status.
	Status state.Status `json:"status"`
	// Pid is the container's init process PID.
	Pid int `json:"pid,omitempty"`
	// Bundle is the path to the container bundle.
	Bundle string `json:"bundle"`
	// Annotations are container annotations.
	Annotations map[string]string `json:"annotations,omitempty"`
}

// Run executes a list of hooks.
func Run(hooks []oci.Hook, hookState *HookState) error {
	if len(hooks) == 0 {
		return nil
	}

	stateJSON, err := json.Marshal(hookState)
	if err != nil {
		return fmt.Errorf("marshal hook state: %w", err)
	}

	for i, hook := range hooks {
		if err := runHook(hook, stateJSON); err != nil {
			return fmt.Errorf("hook[%d] %s: %w", i, hook.Path, err)
		}
	}

	return nil
}

// runHook executes a single hook.
func runHook(hook oci.Hook, stateJSON []byte) error {
	timeout := DefaultTimeout
	if hook.Timeout != nil && *hook.Timeout > 0 {
		timeout = *hook.Timeout
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	args := hook.Args
	if len(args) == 0 {
		args = []string{hook.Path}
	}

	log.WithFields(map[string]interface{}{
		"path":    hook.Path,
		"timeout": timeout,
	}).Debug("executing hook")

	cmd := exec.CommandContext(ctx, hook.Path, args[1:]...)
	cmd.Env = hook.Env
	cmd.Stdin = bytes.NewReader(stateJSON)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("hook timed out after %d seconds", timeout)
		}
		return fmt.Errorf("hook failed: %w (stderr: %s)", err, stderr.String())
	}

	if stdout.Len() > 0 {
		log.WithField("path", hook.Path).Debugf("hook stdout: %s", stdout.String())
	}

	return nil
}

// RunPrestart runs prestart hooks (deprecated but still supported).
func RunPrestart(hooks *oci.Hooks, hookState *HookState) error {
	if hooks == nil {
		return nil
	}
	return Run(hooks.Prestart, hookState)
}

// RunCreateRuntime runs createRuntime hooks.
func RunCreateRuntime(hooks *oci.Hooks, hookState *HookState) error {
	if hooks == nil {
		return nil
	}
	return Run(hooks.CreateRuntime, hookState)
}

// RunCreateContainer runs createContainer hooks.
func RunCreateContainer(hooks *oci.Hooks, hookState *HookState) error {
	if hooks == nil {
		return nil
	}
	return Run(hooks.CreateContainer, hookState)
}

// RunStartContainer runs startContainer hooks.
func RunStartContainer(hooks *oci.Hooks, hookState *HookState) error {
	if hooks == nil {
		return nil
	}
	return Run(hooks.StartContainer, hookState)
}

// RunPoststart runs poststart hooks.
func RunPoststart(hooks *oci.Hooks, hookState *HookState) error {
	if hooks == nil {
		return nil
	}
	return Run(hooks.Poststart, hookState)
}

// RunPoststop runs poststop hooks.
func RunPoststop(hooks *oci.Hooks, hookState *HookState) error {
	if hooks == nil {
		return nil
	}
	return Run(hooks.Poststop, hookState)
}
