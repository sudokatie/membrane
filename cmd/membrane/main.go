// Membrane - OCI Container Runtime
//
// A minimal container runtime implementing the OCI runtime specification.
// Uses Linux namespaces, cgroups v2, and seccomp for process isolation.

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/sudokatie/membrane/internal/container"
	"github.com/sudokatie/membrane/internal/log"
	"github.com/sudokatie/membrane/pkg/oci"
)

const version = "0.1.0"

// Exit codes per OCI runtime spec
const (
	ExitSuccess         = 0
	ExitError           = 1
	ExitContainerError  = 125 // container failed to run
	ExitCommandNotFound = 127 // command not found in container
)

var (
	stateRoot string
	force     bool
	logLevel  string
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(ExitError)
	}
}

func getManager() *container.Manager {
	config := container.DefaultConfig()
	if stateRoot != "" {
		config.StateRoot = stateRoot
	}
	return container.NewManager(config)
}

var rootCmd = &cobra.Command{
	Use:   "membrane",
	Short: "OCI container runtime",
	Long:  "Membrane is a minimal OCI-compliant container runtime using Linux namespaces and cgroups v2.",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		log.SetLevel(log.ParseLevel(logLevel))
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		info := map[string]string{
			"version":    version,
			"ociVersion": oci.Version,
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(info)
	},
}

var createCmd = &cobra.Command{
	Use:   "create <container-id> <bundle>",
	Short: "Create a container",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr := getManager()
		_, err := mgr.Create(&container.CreateOptions{
			ID:     args[0],
			Bundle: args[1],
		})
		if err != nil {
			return exitError(ExitContainerError, err)
		}
		return nil
	},
}

var startCmd = &cobra.Command{
	Use:   "start <container-id>",
	Short: "Start a created container",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr := getManager()
		if err := mgr.Start(&container.StartOptions{
			ID: args[0],
		}); err != nil {
			return exitError(ExitContainerError, err)
		}
		return nil
	},
}

var runCmd = &cobra.Command{
	Use:   "run <container-id> <bundle>",
	Short: "Create and start a container",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr := getManager()
		if err := mgr.Run(&container.CreateOptions{
			ID:     args[0],
			Bundle: args[1],
		}); err != nil {
			return exitError(ExitContainerError, err)
		}
		return nil
	},
}

var stateCmd = &cobra.Command{
	Use:   "state <container-id>",
	Short: "Query container state",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr := getManager()
		st, err := mgr.State(args[0])
		if err != nil {
			return exitError(ExitError, err)
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(st)
	},
}

var killCmd = &cobra.Command{
	Use:   "kill <container-id> [signal]",
	Short: "Send signal to container",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		signal := "SIGTERM"
		if len(args) > 1 {
			signal = args[1]
		}
		mgr := getManager()
		if err := mgr.Kill(&container.KillOptions{
			ID:     args[0],
			Signal: signal,
		}); err != nil {
			return exitError(ExitError, err)
		}
		return nil
	},
}

var deleteCmd = &cobra.Command{
	Use:   "delete <container-id>",
	Short: "Delete a container",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr := getManager()
		if err := mgr.Delete(args[0], force); err != nil {
			return exitError(ExitError, err)
		}
		return nil
	},
}

var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "List containers",
	Aliases: []string{"ls"},
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr := getManager()
		containers, err := mgr.List()
		if err != nil {
			return exitError(ExitError, err)
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "ID\tSTATUS\tPID\tBUNDLE")
		for _, c := range containers {
			fmt.Fprintf(w, "%s\t%s\t%d\t%s\n",
				c.ID, c.State.Status, c.State.Pid, c.Bundle)
		}
		return w.Flush()
	},
}

var specCmd = &cobra.Command{
	Use:   "spec",
	Short: "Generate a default OCI spec",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Output default spec
		spec := map[string]interface{}{
			"ociVersion": oci.Version,
			"root": map[string]interface{}{
				"path":     "rootfs",
				"readonly": false,
			},
			"process": map[string]interface{}{
				"terminal": false,
				"user":     map[string]int{"uid": 0, "gid": 0},
				"args":     []string{"/bin/sh"},
				"env":      []string{"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin", "TERM=xterm"},
				"cwd":      "/",
			},
			"hostname": "container",
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(spec)
	},
}

var execCmd = &cobra.Command{
	Use:   "exec <container-id> <command> [args...]",
	Short: "Execute a command in a running container",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, _ := cmd.Flags().GetString("cwd")
		env, _ := cmd.Flags().GetStringArray("env")
		uid, _ := cmd.Flags().GetUint32("user")
		gid, _ := cmd.Flags().GetUint32("group")

		mgr := getManager()
		user := &oci.User{
			UID: uid,
			GID: gid,
		}
		if uid == 0 && gid == 0 {
			user = nil // use defaults
		}

		if err := mgr.Exec(&container.ExecOptions{
			ID:   args[0],
			Args: args[1:],
			Env:  env,
			Cwd:  cwd,
			User: user,
		}); err != nil {
			return exitError(ExitContainerError, err)
		}
		return nil
	},
}

var waitCmd = &cobra.Command{
	Use:   "wait <container-id>",
	Short: "Wait for container to exit and return exit code",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr := getManager()
		exitCode, err := mgr.Wait(args[0])
		if err != nil {
			return exitError(ExitError, err)
		}
		fmt.Println(exitCode)
		if exitCode != 0 {
			os.Exit(exitCode)
		}
		return nil
	},
}

// exitError wraps an error with an exit code.
type exitCodeError struct {
	code int
	err  error
}

func (e *exitCodeError) Error() string {
	return e.err.Error()
}

func (e *exitCodeError) ExitCode() int {
	return e.code
}

func exitError(code int, err error) error {
	return &exitCodeError{code: code, err: err}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&stateRoot, "root", "", "state directory (default: /run/membrane)")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "log level (error, warn, info, debug)")

	deleteCmd.Flags().BoolVarP(&force, "force", "f", false, "force delete running container")

	execCmd.Flags().String("cwd", "/", "working directory")
	execCmd.Flags().StringArray("env", nil, "environment variables")
	execCmd.Flags().Uint32("user", 0, "user ID")
	execCmd.Flags().Uint32("group", 0, "group ID")

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(stateCmd)
	rootCmd.AddCommand(killCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(specCmd)
	rootCmd.AddCommand(execCmd)
	rootCmd.AddCommand(waitCmd)
}
